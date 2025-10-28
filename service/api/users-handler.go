package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
)

// searchUser searches for users by username pattern
func (rt *_router) searchUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	searchQuery := r.URL.Query().Get("searcheduser")
	if searchQuery == "" {
		ctx.Logger.Error("Search query parameter 'searcheduser' is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Query parameter 'searcheduser' is required"})
		return
	}

	ctx.Logger.WithField("search_query", searchQuery).Info("Searching users")

	users, err := rt.db.SearchUser(searchQuery)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error searching users")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to search users"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}

// setMyUserName updates the authenticated user's username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Username == "" {
		ctx.Logger.Error("New username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("new_username", req.Username).Info("Updating username")

	user, err := rt.db.SetMyUserName(userID, req.Username)
	if err != nil {
		if err.Error() == "username already taken" {
			ctx.Logger.WithError(err).Error("Username already taken")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username already taken"})
			return
		}

		if err.Error() == "username must be between 3 and 25 characters" ||
			err.Error() == "username can only contain letters, numbers, underscores, and hyphens" {
			ctx.Logger.WithError(err).Error("Username validation failed")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ctx.Logger.WithError(err).Error("Error updating username")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update username"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

// setMyPhoto updates the authenticated user's profile picture
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
		Pic string `json:"pic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Pic == "" {
		ctx.Logger.Error("Profile picture is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Profile picture is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).Info("Updating profile picture")

	user, err := rt.db.SetMyPhoto(userID, req.Pic)
	if err != nil {
		if err.Error() == "invalid base64 photo data" {
			ctx.Logger.WithError(err).Error("Invalid photo format")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid photo format. Photo must be base64 encoded"})
			return
		}

		ctx.Logger.WithError(err).Error("Error updating profile picture")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update profile picture"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}
