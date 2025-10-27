package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
	"github.com/val7e/wasaText/service/models"
)

// doLogin handles user login/registration
// If the user doesn't exist, a new user is created with a default profile picture
// Returns the user ID and user profile
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req struct {
		Name models.Username `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate username
	if req.Name == "" {
		ctx.Logger.Error("Username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username is required"})
		return
	}

	ctx.Logger.WithField("username", req.Name).Info("Processing login")

	// Perform login/registration
	user, userID, err := rt.db.DoLogin(req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error during login")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to login"})
		return
	}

	// Prepare response
	response := struct {
		Identifier int64       `json:"identifier"`
		Profile    models.User `json:"profile"`
	}{
		Identifier: userID,
		Profile:    *user,
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// authenticateRequest extracts and validates the Bearer token from the Authorization header
// Returns the authenticated user ID and username, or sends an error response and returns false
func (rt *_router) authenticateRequest(w http.ResponseWriter, r *http.Request, ctx reqcontext.RequestContext) (int64, models.Username, bool) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		ctx.Logger.Error("Missing Authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Missing Authorization header"})
		return 0, "", false
	}

	// Check Bearer format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.Logger.Error("Invalid Authorization header format")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid Authorization format. Expected 'Bearer <token>'"})
		return 0, "", false
	}

	// Parse user ID from token
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid token format")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
		return 0, "", false
	}

	// Verify user exists
	user, err := rt.db.GetUserByID(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("User not found")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token: user not found"})
		return 0, "", false
	}

	ctx.Logger.WithField("user_id", userID).WithField("username", user.Username).Info("User authenticated")
	return userID, user.Username, true
}
