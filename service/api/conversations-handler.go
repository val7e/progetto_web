package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
)

// getMyConversations retrieves all conversations for the authenticated user
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	ctx.Logger.WithField("user_id", userID).Info("Fetching conversations")

	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error fetching conversations")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve conversations"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(conversations)
}

// getConversation retrieves a specific conversation with messages
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	ctx.Logger.WithField("conversation_id", conversationID).WithField("user_id", userID).Info("Fetching conversation")

	// Pass userID to check if user is participant
	conversation, err := rt.db.GetConversation(conversationID, userID)
	if err != nil {
		if err.Error() == "conversation not found" {
			ctx.Logger.WithError(err).Error("Conversation not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Conversation not found"})
			return
		}

		if err.Error() == "user not participant in conversation" {
			ctx.Logger.WithError(err).Error("User not participant")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a participant in this conversation"})
			return
		}

		ctx.Logger.WithError(err).Error("Error fetching conversation")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve conversation"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(conversation)
}

// startConversation creates a new direct conversation
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Parse request body
	var req struct {
		Recipient string `json:"recipient"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Recipient == "" {
		ctx.Logger.Error("Recipient username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient username is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("recipient", req.Recipient).Info("Starting conversation")

	conversation, err := rt.db.StartConversation(userID, req.Recipient)
	if err != nil {
		if err.Error() == "recipient user not found" {
			ctx.Logger.WithError(err).Error("Recipient user not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Recipient user not found"})
			return
		}

		ctx.Logger.WithError(err).Error("Error starting conversation")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to start conversation"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(conversation)
}
