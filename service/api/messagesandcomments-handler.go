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
// User must be a participant in the conversation
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse conversation ID from URL
	conversationID, err := strconv.ParseInt(ps.ByName("conversation_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid conversation ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid conversation ID"})
		return
	}

	// Parse request body
	var req models.NewMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate message type
	if req.Type != "text" && req.Type != "photo" {
		ctx.Logger.Error("Invalid message type")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message type must be 'text' or 'photo'"})
		return
	}

	// Validate content based on type
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

	// Send message in database
	message, err := rt.db.SendMessage(models.Id(conversationID), userID, req)
	if err != nil {
		// Check if user not participant
		if err.Error() == "user not participant in conversation" {
			ctx.Logger.WithError(err).Error("User not participant in conversation")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in this conversation"})
			return
		}

		// Check for invalid base64
		if err.Error() == "invalid base64 photo data" {
			ctx.Logger.WithError(err).Error("Invalid photo format")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid photo format. Photo must be base64 encoded"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error sending message")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send message"})
		return
	}

	// Return created message
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(message)
}

// forwardMessage forwards an existing message to another conversation
// User must be a participant in both conversations
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse message ID from URL (we don't need conversation_id)
	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	// Parse request body
	var req struct {
		RecipientConversationID models.Id `json:"recipient_conversation_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate recipient conversation ID
	if req.RecipientConversationID == 0 {
		ctx.Logger.Error("Recipient conversation ID is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient conversation ID is required"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).WithField("recipient_conversation_id", req.RecipientConversationID).WithField("user_id", userID).Info("Forwarding message")

	// Forward message in database
	forwardedMessage, err := rt.db.ForwardMessage(models.Id(messageID), req.RecipientConversationID, userID)
	if err != nil {
		// Check if user not participant in recipient conversation
		if err.Error() == "user not participant in recipient conversation" {
			ctx.Logger.WithError(err).Error("User not participant in recipient conversation")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in the recipient conversation"})
			return
		}

		// Check if original message not found
		if err.Error() == "original message not found" {
			ctx.Logger.WithError(err).Error("Original message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Original message not found"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error forwarding message")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to forward message"})
		return
	}

	// Return forwarded message
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(forwardedMessage)
}

// deleteMessage deletes a message from a conversation
// Only the sender can delete their own messages
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse conversation ID and message ID from URL
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

	// Delete message in database
	err = rt.db.DeleteMessage(models.Id(messageID), models.Id(conversationID), userID)
	if err != nil {
		// Check if message not found
		if err.Error() == "message not found" {
			ctx.Logger.WithError(err).Error("Message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message not found"})
			return
		}

		// Check if user not authorized (not the sender)
		if err.Error() == "unauthorized: user is not the sender" {
			ctx.Logger.WithError(err).Error("User not authorized to delete message")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You can only delete your own messages"})
			return
		}

		// Check if message doesn't belong to conversation
		if err.Error() == "message does not belong to specified conversation" {
			ctx.Logger.WithError(err).Error("Message doesn't belong to conversation")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message does not belong to this conversation"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error deleting message")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete message"})
		return
	}

	// Return success (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}

// commentMessage adds a comment to a message
// User must be a participant in the conversation
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse conversation ID and message ID from URL
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
	var req models.NewComment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate comment text
	if req.Text == "" {
		ctx.Logger.Error("Comment text is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Comment text is required"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).WithField("user_id", userID).Info("Adding comment to message")

	// Add comment in database
	comment, err := rt.db.CommentMessage(models.Id(messageID), models.Id(conversationID), userID, req)
	if err != nil {
		// Check if message not found
		if err.Error() == "message not found" {
			ctx.Logger.WithError(err).Error("Message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message not found"})
			return
		}

		// Check if user not participant
		if err.Error() == "user not participant in conversation" {
			ctx.Logger.WithError(err).Error("User not participant in conversation")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in this conversation"})
			return
		}

		// Check if message doesn't belong to conversation
		if err.Error() == "message does not belong to specified conversation" {
			ctx.Logger.WithError(err).Error("Message doesn't belong to conversation")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message does not belong to this conversation"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error adding comment")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add comment"})
		return
	}

	// Return created comment
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(comment)
}

// uncommentMessage removes the user's comment from a message
// User can only remove their own comment
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse conversation ID and message ID from URL
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

	// Remove comment in database
	err = rt.db.UncommentMessage(models.Id(messageID), models.Id(conversationID), userID)
	if err != nil {
		// Check if message not found
		if err.Error() == "message not found" {
			ctx.Logger.WithError(err).Error("Message not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message not found"})
			return
		}

		// Check if comment not found
		if err.Error() == "comment not found or user is not the author" {
			ctx.Logger.WithError(err).Error("Comment not found or user not author")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Comment not found or you are not the author"})
			return
		}

		// Check if message doesn't belong to conversation
		if err.Error() == "message does not belong to specified conversation" {
			ctx.Logger.WithError(err).Error("Message doesn't belong to conversation")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Message does not belong to this conversation"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error removing comment")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to remove comment"})
		return
	}

	// Return success (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}

// getComments retrieves all comments for a specific message
// User must be a participant in the conversation to view comments
func (rt *_router) getComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse message ID from URL
	messageID, err := strconv.ParseInt(ps.ByName("message_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid message ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid message ID"})
		return
	}

	ctx.Logger.WithField("message_id", messageID).WithField("user_id", userID).Info("Fetching comments")

	// Get comments from database
	comments, err := rt.db.GetComments(models.Id(messageID))
	if err != nil {
		ctx.Logger.WithError(err).Error("Error fetching comments")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve comments"})
		return
	}

	// Return comments
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(comments)
}
