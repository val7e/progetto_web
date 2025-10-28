package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
)

// doLogin handles user registration/login
// if the user does not exist, it will be registered with a default profile picture
// If the user exists, it returns the existing user profile
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Username == "" {
		ctx.Logger.Error("Username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username is required"})
		return
	}

	ctx.Logger.WithField("username", req.Username).Info("Login/Registration attempt")

	user, isNewUser, err := rt.db.DoLogin(req.Username)
	if err != nil {
		if err.Error() == "username must be between 3 and 25 characters" ||
			err.Error() == "username can only contain letters, numbers, underscores, and hyphens" {
			ctx.Logger.WithError(err).Error("Username validation failed")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ctx.Logger.WithError(err).Error("Error during login/registration")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to login/register"})
		return
	}

	// Return user identifier
	// used by the client in Authorization header
	response := map[string]interface{}{
		"identifier": strconv.FormatInt(user.Id, 10),
		"username":   user.Username,
		"pic":        user.Pic,
	}

	if isNewUser {
		ctx.Logger.WithField("user_id", user.Id).Info("New user registered")
		w.WriteHeader(http.StatusCreated)
	} else {
		ctx.Logger.WithField("user_id", user.Id).Info("Existing user logged in")
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(response)
}

// getUserFromAuth extracts user ID from Authorization header
// What to write: "Bearer <user_id>"
func (rt *_router) getUserFromAuth(r *http.Request) (int64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("authorization header required")
	}

	// Split "Bearer <user_id>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, fmt.Errorf("invalid authorization format. Expected: Bearer <user_id>")
	}

	// Parse user ID
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user identifier")
	}

	return userID, nil
}
