package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
	"github.com/val7e/wasaText/service/models"
)

// searchUser searches for users by username
// Returns a list of usernames matching the search query
func (rt *_router) searchUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Get search query from URL parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		ctx.Logger.Error("Search query is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Search query parameter 'q' is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("query", query).Info("Searching users")

	// Search users in database
	users, err := rt.db.SearchUsers(models.Username(query))
	if err != nil {
		ctx.Logger.WithError(err).Error("Error searching users")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to search users"})
		return
	}

	// Return users list
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}

// setMyUserName updates the authenticated user's username
// The username in the URL must match the authenticated user
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, currentUsername, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Get username from URL parameter
	urlUsername := models.Username(ps.ByName("username"))

	// Verify the URL username matches the authenticated user
	if urlUsername != currentUsername {
		ctx.Logger.Error("Username in URL does not match authenticated user")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "You can only update your own username"})
		return
	}

	// Parse request body
	var req struct {
		Username models.Username `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate new username
	if req.Username == "" {
		ctx.Logger.Error("Username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("old_username", currentUsername).WithField("new_username", req.Username).Info("Updating username")

	// Update username in database
	user, err := rt.db.SetMyUserName(userID, req.Username)
	if err != nil {
		// Check if username already taken
		if err.Error() == "username already exists" {
			ctx.Logger.WithError(err).Error("Username already taken")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username already taken"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error updating username")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update username"})
		return
	}

	// Return updated user
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

// setMyPic updates the authenticated user's profile picture
// The username in the URL must match the authenticated user
func (rt *_router) setMyPic(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, currentUsername, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Get username from URL parameter
	urlUsername := models.Username(ps.ByName("username"))

	// Verify the URL username matches the authenticated user
	if urlUsername != currentUsername {
		ctx.Logger.Error("Username in URL does not match authenticated user")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "You can only update your own profile picture"})
		return
	}

	// Parse request body
	var req struct {
		Pic models.Pic `json:"pic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate profile picture (base64 encoded)
	if req.Pic == "" {
		ctx.Logger.Error("Profile picture is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Profile picture is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("username", currentUsername).Info("Updating profile picture")

	// Update profile picture in database
	user, err := rt.db.SetMyPhoto(userID, string(req.Pic))
	if err != nil {
		// Check for invalid base64
		if err.Error() == "invalid base64 photo data" {
			ctx.Logger.WithError(err).Error("Invalid photo format")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid photo format. Photo must be base64 encoded"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error updating profile picture")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update profile picture"})
		return
	}

	// Return updated user
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}
