package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
	"github.com/val7e/wasaText/service/models"
)

// sendMessage sends a new message (text or photo) in a conversation
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	conversationID, err := strconv.ParseInt(ps.ByName("conversation_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid conversation ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid conversation ID"})
		return
	}

	// Parse request body
	var req struct {
		Type  string  `json:"type"`
		Text  *string `json:"text,omitempty"`
		Photo *string `json:"photo,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Type != "text" && req.Type != "photo" {
		ctx.Logger.Error("Invalid message type")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message type must be 'text' or 'photo'"})
		return
	}

	if req.Type == "text" && (req.Text == nil || *req.Text == "") {
		ctx.Logger.Error("Text message requires text content")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Text message requires text content"})
		return
	}

	if req.Type == "photo" && (req.Photo == nil || *req.Photo == "") {
		ctx.Logger.Error("Photo message requires photo content")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Photo message requires photo content (base64 encoded)"})
		return
	}

	ctx.Logger.WithField("conversation_id", conversationID).WithField("user_id", userID).WithField("type", req.Type).Info("Sending message")

	newMsg := models.NewMessage{
		Type:  req.Type,
		Text:  req.Text,
		Photo: req.Photo,
	}

	message, err := rt.db.SendMessage(conversationID, userID, newMsg)
	if err != nil {
		if err.Error() == "user not participant in conversation" {
			ctx.Logger.WithError(err).Error("User not participant in conversation")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in this conversation"})
			return
		}

		if err.Error() == "invalid base64 photo data" {
			ctx.Logger.WithError(err).Error("Invalid photo format")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid photo format. Photo must be base64 encoded"})
			return
		}

		ctx.Logger.WithError(err).Error("Error sending message")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send message"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(message)
}

// forwardMessage forwards a message to another conversation
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	// Parse request body
	var req struct {
		RecipientUsername string `json:"recipient_username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.RecipientUsername == "" {
		ctx.Logger.Error("Recipient username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient username is required"})
		return
	}

	// Use StartConversation to get or create the conversation with the recipient
	conversation, err := rt.db.StartConversation(userID, req.RecipientUsername)
	if err != nil {
		if err.Error() == "recipient user not found" {
			ctx.Logger.WithError(err).Error("Recipient user not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient user not found"})
			return
		}
		ctx.Logger.WithError(err).Error("Error getting/creating conversation")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get conversation"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).WithField("recipient_username", req.RecipientUsername).WithField("user_id", userID).Info("Forwarding message")

	forwardedMessage, err := rt.db.ForwardMessage(messageID, conversation.Id, userID)
	if err != nil {
		if err.Error() == "original message not found" {
			ctx.Logger.WithError(err).Error("Original message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Original message not found"})
			return
		}

		ctx.Logger.WithError(err).Error("Error forwarding message")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to forward message"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(forwardedMessage)
}

// deleteMessage deletes a message from a conversation
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	conversationID, err := strconv.ParseInt(ps.ByName("conversation_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid conversation ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid conversation ID"})
		return
	}

	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	ctx.Logger.WithField("conversation_id", conversationID).WithField("message_id", messageID).WithField("user_id", userID).Info("Deleting message")

	err = rt.db.DeleteMessage(messageID, conversationID, userID)
	if err != nil {
		if err.Error() == "message not found" {
			ctx.Logger.WithError(err).Error("Message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message not found"})
			return
		}

		if err.Error() == "unauthorized: user is not the sender" {
			ctx.Logger.WithError(err).Error("User not authorized to delete message")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You can only delete your own messages"})
			return
		}

		if err.Error() == "message does not belong to specified conversation" {
			ctx.Logger.WithError(err).Error("Message doesn't belong to conversation")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message does not belong to this conversation"})
			return
		}

		ctx.Logger.WithError(err).Error("Error deleting message")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete message"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// commentMessage adds a comment to a message
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	conversationID, err := strconv.ParseInt(ps.ByName("conversation_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid conversation ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid conversation ID"})
		return
	}

	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	// Parse request body
	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Text == "" {
		ctx.Logger.Error("Comment text is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Comment text is required"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).WithField("user_id", userID).Info("Adding comment to message")

	newComment := models.NewComment{
		Text: req.Text,
	}

	comment, err := rt.db.CommentMessage(messageID, conversationID, userID, newComment)
	if err != nil {
		if err.Error() == "message not found" {
			ctx.Logger.WithError(err).Error("Message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message not found"})
			return
		}

		if err.Error() == "user not participant in conversation" {
			ctx.Logger.WithError(err).Error("User not participant in conversation")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in this conversation"})
			return
		}

		if err.Error() == "message does not belong to specified conversation" {
			ctx.Logger.WithError(err).Error("Message doesn't belong to conversation")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message does not belong to this conversation"})
			return
		}

		ctx.Logger.WithError(err).Error("Error adding comment")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add comment"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(comment)
}

// uncommentMessage removes the user's comment from a message
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	conversationID, err := strconv.ParseInt(ps.ByName("conversation_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid conversation ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid conversation ID"})
		return
	}

	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).WithField("user_id", userID).Info("Removing comment from message")

	err = rt.db.UncommentMessage(messageID, conversationID, userID)
	if err != nil {
		if err.Error() == "message not found" {
			ctx.Logger.WithError(err).Error("Message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message not found"})
			return
		}

		if err.Error() == "comment not found or user is not the author" {
			ctx.Logger.WithError(err).Error("Comment not found or user not author")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Comment not found or you are not the author"})
			return
		}

		if err.Error() == "message does not belong to specified conversation" {
			ctx.Logger.WithError(err).Error("Message doesn't belong to conversation")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message does not belong to this conversation"})
			return
		}

		ctx.Logger.WithError(err).Error("Error removing comment")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to remove comment"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getComments retrieves all comments for a specific message
func (rt *_router) getComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).Info("Fetching comments")

	comments, err := rt.db.GetComments(messageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error fetching comments")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve comments"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(comments)
}
