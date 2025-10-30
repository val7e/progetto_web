package api

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/julienschmidt/httprouter"
    "github.com/val7e/wasaText/service/api/reqcontext"
    "github.com/val7e/wasaText/service/database"
)

// createGroup creates a new group
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Name == "" {
		ctx.Logger.Error("Group name is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group name is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("group_name", req.Name).Info("Creating group")

	group, err := rt.db.CreateGroup(userID, req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error creating group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create group"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(group)
}

// getGroup retrieves group information including members
func (rt *_router) getGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("group_id", groupID).Info("Getting group")

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		if err.Error() == database.ErrGroupNotFound {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		ctx.Logger.WithError(err).Error("Error getting group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get group"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// setGroupName updates a group's name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Name == "" {
		ctx.Logger.Error("Group name is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group name is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("group_id", groupID).WithField("new_name", req.Name).Info("Updating group name")

    group, err := rt.db.SetGroupName(groupID, req.Name)
    if err != nil {
		if err.Error() == database.ErrGroupNotFound {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		ctx.Logger.WithError(err).Error("Error updating group name")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update group name"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// getGroupByConversation returns the group bound to a conversation id (only for participants)
// getGroupByConversation removed to avoid backend changes per user request

// setGroupPhoto updates a group's photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	var req struct {
		Photo string `json:"photo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Photo == "" {
		ctx.Logger.Error("Group photo is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group photo is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("group_id", groupID).Info("Updating group photo")
	group, err := rt.db.SetGroupPhoto(groupID, req.Photo)
	if err != nil {
		if err.Error() == database.ErrGroupNotFound {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		if err.Error() == "invalid base64 photo data" || err.Error() == "invalid base64 photo data: illegal base64 data at input byte 0" {
			ctx.Logger.WithError(err).Error("Invalid photo format")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid photo format. Photo must be base64 encoded"})
			return
		}

		ctx.Logger.WithError(err).Error("Error updating group photo")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update group photo"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// addToGroup adds members to a group
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	var req struct {
		Members []string `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if len(req.Members) == 0 {
		ctx.Logger.Error("At least one member username is required")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "At least one member username is required"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("group_id", groupID).WithField("members", req.Members).Info("Adding members to group")

	group, err := rt.db.AddToGroup(groupID, req.Members)
	if err != nil {
		if err.Error() == database.ErrGroupNotFound {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		ctx.Logger.WithError(err).Error("Error adding members to group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add members to group"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(group)
}

// leaveGroup removes the authenticated user from a group
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from Authorization header
	userID, err := rt.getUserFromAuth(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Authorization failed")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	groupID, err := strconv.ParseInt(ps.ByName("group_id"), 10, 64)
	if err != nil {
		ctx.Logger.WithError(err).Error("Invalid group ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	ctx.Logger.WithField("user_id", userID).WithField("group_id", groupID).Info("User leaving group")

	err = rt.db.LeaveGroup(groupID, userID)
	if err != nil {
		if err.Error() == database.ErrGroupNotFound {
			ctx.Logger.WithError(err).Error("Group not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}

		if err.Error() == "user not member of group" {
			ctx.Logger.WithError(err).Error("User not in group")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "You are not a member of this group"})
			return
		}

		ctx.Logger.WithError(err).Error("Error leaving group")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to leave group"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
