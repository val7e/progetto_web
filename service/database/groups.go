package database

import (
    "database/sql"
    "encoding/base64"
    "errors"
    "fmt"

    "github.com/val7e/wasaText/service/models"
)

const ErrGroupNotFound = "group not found"

// In this refactor, a Group is just a Conversation of type 'group'.
// groupID corresponds to conversation.id

// CreateGroup creates a new conversation of type 'group', sets optional name, and adds creator as participant.
func (db *appdbimpl) CreateGroup(creatorID int64, name string) (*models.Group, error) {
    // Create conversation
    res, err := db.c.Exec("INSERT INTO conversations (type, name, created_at, updated_at) VALUES ('group', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", name)
    if err != nil {
        return nil, fmt.Errorf("error creating conversation: %w", err)
    }
    convID, err := res.LastInsertId()
    if err != nil {
        return nil, fmt.Errorf("error getting conversation ID: %w", err)
    }

    // Add creator as participant
    if _, err := db.c.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", convID, creatorID); err != nil {
        return nil, fmt.Errorf("error adding creator to conversation: %w", err)
    }

    return db.getGroupByID(convID)
}

// GetGroup retrieves group information by conversation id
func (db *appdbimpl) GetGroup(groupID int64) (*models.Group, error) {
    return db.getGroupByID(groupID)
}

// SetGroupName updates the conversation name
func (db *appdbimpl) SetGroupName(groupID int64, name string) (*models.Group, error) {
    res, err := db.c.Exec("UPDATE conversations SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND type = 'group'", name, groupID)
    if err != nil {
        return nil, fmt.Errorf("error updating group name: %w", err)
    }
    rows, _ := res.RowsAffected()
    if rows == 0 {
        return nil, fmt.Errorf(ErrGroupNotFound)
    }
    return db.getGroupByID(groupID)
}

// SetGroupPhoto updates the conversation picture (base64)
func (db *appdbimpl) SetGroupPhoto(groupID int64, photoBase64 string) (*models.Group, error) {
    if _, err := base64.StdEncoding.DecodeString(photoBase64); err != nil {
        return nil, fmt.Errorf("invalid base64 photo data: %w", err)
    }
    res, err := db.c.Exec("UPDATE conversations SET convo_pic = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND type = 'group'", photoBase64, groupID)
    if err != nil {
        return nil, fmt.Errorf("error updating group photo: %w", err)
    }
    rows, _ := res.RowsAffected()
    if rows == 0 {
        return nil, fmt.Errorf(ErrGroupNotFound)
    }
    return db.getGroupByID(groupID)
}

// AddToGroup adds participants to the group conversation
func (db *appdbimpl) AddToGroup(groupID int64, memberUsernames []string) (*models.Group, error) {
    // Ensure conversation exists and is a group
    var typ string
    if err := db.c.QueryRow("SELECT type FROM conversations WHERE id = ?", groupID).Scan(&typ); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf(ErrGroupNotFound)
        }
        return nil, fmt.Errorf("error checking group: %w", err)
    }
    if typ != "group" {
        return nil, fmt.Errorf(ErrGroupNotFound)
    }

    for _, username := range memberUsernames {
        var userID int64
        err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
        if errors.Is(err, sql.ErrNoRows) {
            continue
        }
        if err != nil {
            return nil, fmt.Errorf("error finding user: %w", err)
        }
        if _, err := db.c.Exec("INSERT OR IGNORE INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", groupID, userID); err != nil {
            return nil, fmt.Errorf("error adding member to conversation: %w", err)
        }
    }

    return db.getGroupByID(groupID)
}

// LeaveGroup removes the user from conversation participants
func (db *appdbimpl) LeaveGroup(groupID, userID int64) error {
    res, err := db.c.Exec("DELETE FROM conversation_participants WHERE conversation_id = ? AND user_id = ?", groupID, userID)
    if err != nil {
        return fmt.Errorf("error leaving group: %w", err)
    }
    rows, _ := res.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("user not member of group")
    }
    return nil
}

// Helper function to assemble Group from conversation and participants
func (db *appdbimpl) getGroupByID(groupID int64) (*models.Group, error) {
    var name sql.NullString
    var typ string
    var convoPic sql.NullString
    err := db.c.QueryRow("SELECT name, type, convo_pic FROM conversations WHERE id = ?", groupID).Scan(&name, &typ, &convoPic)
    if errors.Is(err, sql.ErrNoRows) || typ != "group" {
        return nil, fmt.Errorf(ErrGroupNotFound)
    }
    if err != nil {
        return nil, fmt.Errorf("error getting group: %w", err)
    }

    // Members
    rows, err := db.c.Query(`
        SELECT u.username
        FROM users u
        INNER JOIN conversation_participants cp ON u.id = cp.user_id
        WHERE cp.conversation_id = ?
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

    var photoPtr *string
    if convoPic.Valid {
        v := convoPic.String
        photoPtr = &v
    }

    return &models.Group{
        Id:         groupID,
        Name:       name.String,
        Members:    members,
        GroupPhoto: photoPtr,
    }, nil
}
