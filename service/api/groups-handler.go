package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/val7e/wasaText/service/api/reqcontext"
	"github.com/val7e/wasaText/service/models"
)

// createGroup creates a new group with the authenticated user as creator
// The creator is automatically added as the first member
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse request body
	var req struct {
		Name models.Name `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate group name
	if req.Name == "" {
		ctx.Logger.Error("Group name is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group name is required"})
		return
	}

	ctx.Logger.WithField("creator_id", userID).WithField("group_name", req.Name).Info("Creating group")

	// Create group in database
	group, err := rt.db.CreateGroup(userID, req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error creating group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create group"})
		return
	}

	// Return created group
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(group)
}

// setGroupName updates a group's name
// Only group members can update the name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse group ID from URL
	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Parse request body
	var req struct {
		Name models.Name `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate group name
	if req.Name == "" {
		ctx.Logger.Error("Group name is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group name is required"})
		return
	}

	ctx.Logger.WithField("group_id", groupID).WithField("user_id", userID).Info("Updating group name")

	// Update group name in database
	group, err := rt.db.SetGroupName(models.Id(groupID), req.Name)
	if err != nil {
		// Check if group not found
		if err.Error() == "group not found" {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error updating group name")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update group name"})
		return
	}

	// Return updated group
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// setGroupPhoto updates a group's photo
// Only group members can update the photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse group ID from URL
	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Parse request body
	var req struct {
		Photo models.Pic `json:"photo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate photo (base64 encoded)
	if req.Photo == "" {
		ctx.Logger.Error("Photo is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Photo is required"})
		return
	}

	ctx.Logger.WithField("group_id", groupID).WithField("user_id", userID).Info("Updating group photo")

	// Update group photo in database
	group, err := rt.db.SetGroupPhoto(models.Id(groupID), string(req.Photo))
	if err != nil {
		// Check if group not found
		if err.Error() == "group not found" {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
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
		ctx.Logger.WithError(err).Error("Error updating group photo")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update group photo"})
		return
	}

	// Return updated group
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// addToGroup adds members to a group by their usernames
// Only group members can add new members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse group ID from URL
	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Parse request body
	var req struct {
		Members []models.Username `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate members list
	if len(req.Members) == 0 {
		ctx.Logger.Error("Members list is empty")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "At least one member username is required"})
		return
	}

	ctx.Logger.WithField("group_id", groupID).WithField("user_id", userID).WithField("members_count", len(req.Members)).Info("Adding members to group")

	// Add members to group in database
	group, err := rt.db.AddToGroup(models.Id(groupID), req.Members)
	if err != nil {
		// Check if group not found
		if err.Error() == "group not found" {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error adding members to group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add members to group"})
		return
	}

	// Return updated group with new members
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// leaveGroup removes the authenticated user from a group
// The user must be a member of the group to leave it
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Authenticate user
	userID, _, ok := rt.authenticateRequest(w, r, ctx)
	if !ok {
		return
	}

	// Parse group ID from URL
	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	ctx.Logger.WithField("group_id", groupID).WithField("user_id", userID).Info("User leaving group")

	// Remove user from group in database
	err = rt.db.LeaveGroup(models.Id(groupID), userID)
	if err != nil {
		// Check if user not member of group
		if err.Error() == "user not member of group" {
			ctx.Logger.WithError(err).Error("User not member of group")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a member of this group"})
			return
		}

		// Other errors
		ctx.Logger.WithError(err).Error("Error leaving group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to leave group"})
		return
	}

	// Return success (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}
