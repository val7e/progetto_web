package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
	"github.com/val7e/wasaText/service/models"
)

// getMyConversations retrieves all conversations for the authenticated user
// Returns a list of conversation summaries with last message preview
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	ctx.Logger.WithField("user_id", userID).Info("Fetching conversations")

	// Get conversations from database
	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error fetching conversations")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve conversations"})
		return
	}

	// Return conversations
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(conversations)
}

// getConversation retrieves a specific conversation with all messages
// Requires the user to be a participant in the conversation
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	ctx.Logger.WithField("conversation_id", conversationID).WithField("user_id", userID).Info("Fetching conversation")

	// Get conversation from database
	conversation, err := rt.db.GetConversation(models.Id(conversationID), userID)
	if err != nil {
		// Check if it's an authorization error (user not participant)
		if err.Error() == "user not participant in conversation" {
			ctx.Logger.WithError(err).Error("User not authorized to view conversation")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in this conversation"})
			return
		}

		// Check if conversation not found
		if err.Error() == "conversation not found" {
			ctx.Logger.WithError(err).Error("Conversation not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Conversation not found"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error fetching conversation")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve conversation"})
		return
	}

	// Return conversation
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(conversation)
}

// startConversation creates a new direct conversation between the authenticated user and a recipient
// If a conversation already exists, it returns the existing one
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse request body
	var req struct {
		Recipient models.Username `json:"recipient"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate recipient username
	if req.Recipient == "" {
		ctx.Logger.Error("Recipient username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient username is required"})
		return
	}

	ctx.Logger.WithField("sender_id", userID).WithField("recipient", req.Recipient).Info("Starting conversation")

	// Start conversation in database
	conversation, err := rt.db.StartConversation(userID, req.Recipient)
	if err != nil {
		// Check if recipient not found
		if err.Error() == "recipient user not found" {
			ctx.Logger.WithError(err).Error("Recipient user not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient user not found"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error starting conversation")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to start conversation"})
		return
	}

	// Return conversation (201 Created for new conversations)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(conversation)
}
