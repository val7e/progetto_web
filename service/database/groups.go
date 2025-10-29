package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/val7e/wasaText/service/models"
)

const ErrGroupNotFound = "group not found"

// CreateGroup creates a new group
func (db *appdbimpl) CreateGroup(creatorID int64, name string) (*models.Group, error) {
	// Create conversation first with type 'group'
	result, err := db.c.Exec(
		"INSERT INTO conversations (type, created_at, updated_at) VALUES ('group', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
	)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation: %w", err)
	}

	conversationID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting conversation ID: %w", err)
	}

	// Insert group with conversation_id
	result, err = db.c.Exec(
		"INSERT INTO groups (name, conversation_id, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		name, conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating group: %w", err)
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting group ID: %w", err)
	}

	// Add creator as participant in conversation
	_, err = db.c.Exec(
		"INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)",
		conversationID, creatorID,
	)
	if err != nil {
		return nil, fmt.Errorf("error adding creator to conversation: %w", err)
	}

	// Add creator as first member
	_, err = db.c.Exec(
		"INSERT INTO group_members (group_id, user_id, joined_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
		groupID, creatorID,
	)
	if err != nil {
		return nil, fmt.Errorf("error adding creator to group: %w", err)
	}

	return db.getGroupByID(groupID)
}

// GetGroup retrieves group information by ID
func (db *appdbimpl) GetGroup(groupID int64) (*models.Group, error) {
	return db.getGroupByID(groupID)
}

// SetGroupName sets a group's name
func (db *appdbimpl) SetGroupName(groupID int64, name string) (*models.Group, error) {
	_, err := db.c.Exec(
		"UPDATE groups SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating group name: %w", err)
	}

	return db.getGroupByID(groupID)
}

// SetGroupPhoto sets a group's photo
func (db *appdbimpl) SetGroupPhoto(groupID int64, photoBase64 string) (*models.Group, error) {
	photoBytes, err := base64.StdEncoding.DecodeString(photoBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 photo data: %w", err)
	}

	// Store as BLOB
	_, err = db.c.Exec(
		"UPDATE groups SET group_photo = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		photoBytes, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating group photo: %w", err)
	}

	return db.getGroupByID(groupID)
}

// AddToGroup adds members to a group
func (db *appdbimpl) AddToGroup(groupID int64, memberUsernames []string) (*models.Group, error) {
	// Get group's conversation_id
	var conversationID int64
	err := db.c.QueryRow("SELECT conversation_id FROM groups WHERE id = ?", groupID).Scan(&conversationID)
	if err != nil {
		return nil, fmt.Errorf(ErrGroupNotFound)
	}

	// Add each member
	for _, username := range memberUsernames {
		var userID int64
		err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
		if errors.Is(err, sql.ErrNoRows) {
			continue // Skip non-existent users
		}
		if err != nil {
			return nil, fmt.Errorf("error finding user: %w", err)
		}

		// Insert member (ignore if already exists)
		_, err = db.c.Exec(
			"INSERT OR IGNORE INTO group_members (group_id, user_id, joined_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
			groupID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("error adding member: %w", err)
		}

		// Add to conversation_participants
		_, err = db.c.Exec(
			"INSERT OR IGNORE INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)",
			conversationID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("error adding member to conversation: %w", err)
		}
	}

	return db.getGroupByID(groupID)
}

// LeaveGroup removes a user from a group
func (db *appdbimpl) LeaveGroup(groupID, userID int64) error {
	result, err := db.c.Exec(
		"DELETE FROM group_members WHERE group_id = ? AND user_id = ?",
		groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("error leaving group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not member of group")
	}

	return nil
}

// Helper function to get group by ID
func (db *appdbimpl) getGroupByID(groupID int64) (*models.Group, error) {
	var group models.Group
	var photoBytes []byte

	err := db.c.QueryRow(
		"SELECT id, name, conversation_id, group_photo FROM groups WHERE id = ?",
		groupID,
	).Scan(&group.Id, &group.Name, &group.ConversationID, &photoBytes)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf(ErrGroupNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting group: %w", err)
	}

	// Convert BLOB to base64 if photo exists
	if len(photoBytes) > 0 {
		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		pic := photoBase64
		group.GroupPhoto = &pic
	}

	// Get members
	rows, err := db.c.Query(`
        SELECT u.username
        FROM users u
        INNER JOIN group_members gm ON u.id = gm.user_id
        WHERE gm.group_id = ?
        ORDER BY u.username
        LIMIT 1000
    `, groupID)

	if err != nil {
		return nil, fmt.Errorf("error getting members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var members []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, fmt.Errorf("error scanning member: %w", err)
		}
		members = append(members, username)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	group.Members = members

	return &group, nil
}
