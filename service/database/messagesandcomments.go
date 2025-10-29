package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/val7e/wasaText/service/models"
)

// SendMessage sends a message in a conversation
func (db *appdbimpl) SendMessage(conversationID, senderID int64, message models.NewMessage) (*models.Message, error) {
	// Verify user is participant in conversation
	var participantCount int
	err := db.c.QueryRow(`
		SELECT COUNT(*) FROM conversation_participants
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, senderID).Scan(&participantCount)

	if err != nil || participantCount == 0 {
		return nil, fmt.Errorf("user not participant in conversation")
	}

	// Handle photo if present
	var photoBytes []byte
	if message.Photo != nil && *message.Photo != "" {
		photoBytes, err = base64.StdEncoding.DecodeString(*message.Photo)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 photo data: %w", err)
		}
	}

	// Handle text
	var text sql.NullString
	if message.Text != nil {
		text = sql.NullString{String: *message.Text, Valid: true}
	}

	// Insert message
	result, err := db.c.Exec(`
		INSERT INTO messages (conversation_id, sender_id, type, text, photo, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, conversationID, senderID, message.Type, text, photoBytes, time.Now())

	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting message ID: %w", err)
	}

	// Return the created message
	return db.getMessageByID(messageID)
}

// ForwardMessage forwards an existing message to another conversation
func (db *appdbimpl) ForwardMessage(messageID, recipientConversationID, authorID int64) (*models.Message, error) {
	// Verify author is participant in recipient conversation
	var participantCount int
	err := db.c.QueryRow(`
		SELECT COUNT(*) FROM conversation_participants
		WHERE conversation_id = ? AND user_id = ?
	`, recipientConversationID, authorID).Scan(&participantCount)

	if err != nil || participantCount == 0 {
		return nil, fmt.Errorf("user not participant in recipient conversation")
	}

	// Get original message
	var text sql.NullString
	var photoBytes []byte
	var msgType string

	err = db.c.QueryRow(`
		SELECT type, text, photo FROM messages WHERE id = ?
	`, messageID).Scan(&msgType, &text, &photoBytes)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("original message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting original message: %w", err)
	}

	// Create forwarded message
	result, err := db.c.Exec(`
		INSERT INTO messages (conversation_id, sender_id, type, text, photo, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, recipientConversationID, authorID, msgType, text, photoBytes, time.Now())

	if err != nil {
		return nil, fmt.Errorf("error forwarding message: %w", err)
	}

	newMessageID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting forwarded message ID: %w", err)
	}

	return db.getMessageByID(newMessageID)
}

// DeleteMessage deletes a message
func (db *appdbimpl) DeleteMessage(messageID, conversationID, userID int64) error {
	// Verify message exists in the specified conversation and user is the sender
	var senderID int64
	var msgConversationID int64

	err := db.c.QueryRow(
		"SELECT sender_id, conversation_id FROM messages WHERE id = ?",
		messageID,
	).Scan(&senderID, &msgConversationID)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("message not found")
	}
	if err != nil {
		return fmt.Errorf("error finding message: %w", err)
	}

	// Verify conversation ID matches
	if msgConversationID != conversationID {
		return fmt.Errorf("message does not belong to specified conversation")
	}

	// Verify user is the sender
	if senderID != userID {
		return fmt.Errorf("unauthorized: user is not the sender")
	}

	// Delete message (and related comments)
	_, err = db.c.Exec("DELETE FROM messages WHERE id = ?", messageID)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}

	return nil
}

// CommentMessage adds a comment to a message
func (db *appdbimpl) CommentMessage(messageID, conversationID, authorID int64, comment models.NewComment) (*models.Comment, error) {
	// Verify message exists and belongs to the conversation
	var msgConversationID int64
	err := db.c.QueryRow(
		"SELECT conversation_id FROM messages WHERE id = ?",
		messageID,
	).Scan(&msgConversationID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error finding message: %w", err)
	}

	if msgConversationID != conversationID {
		return nil, fmt.Errorf("message does not belong to specified conversation")
	}

	// Verify author is participant in conversation
	var participantCount int
	err = db.c.QueryRow(`
		SELECT COUNT(*) FROM conversation_participants
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, authorID).Scan(&participantCount)

	if err != nil || participantCount == 0 {
		return nil, fmt.Errorf("user not participant in conversation")
	}

	// Insert comment
	result, err := db.c.Exec(`
		INSERT INTO comments (message_id, user_id, text, timestamp)
		VALUES (?, ?, ?, ?)
	`, messageID, authorID, comment.Text, time.Now())

	if err != nil {
		return nil, fmt.Errorf("error adding comment: %w", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting comment ID: %w", err)
	}

	// Get username
	var username string
	err = db.c.QueryRow("SELECT username FROM users WHERE id = ?", authorID).Scan(&username)
	if err != nil {
		return nil, fmt.Errorf("error getting username: %w", err)
	}

	return &models.Comment{
		Id:     commentID,
		Author: username,
		Text:   comment.Text,
	}, nil
}

// UncommentMessage deletes a comment from a message
func (db *appdbimpl) UncommentMessage(messageID, conversationID, userID int64) error {
	// Verify message belongs to conversation
	var msgConversationID int64
	err := db.c.QueryRow(
		"SELECT conversation_id FROM messages WHERE id = ?",
		messageID,
	).Scan(&msgConversationID)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("message not found")
	}
	if err != nil {
		return fmt.Errorf("error finding message: %w", err)
	}

	if msgConversationID != conversationID {
		return fmt.Errorf("message does not belong to specified conversation")
	}

	// Find and delete user's comment on this message
	result, err := db.c.Exec(
		"DELETE FROM comments WHERE message_id = ? AND user_id = ?",
		messageID, userID,
	)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found or user is not the author")
	}

	return nil
}

// GetComments retrieves all comments for a message
func (db *appdbimpl) GetComments(messageID int64) ([]models.Comment, error) {
	rows, err := db.c.Query(`
		SELECT c.id, u.username, c.text
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		WHERE c.message_id = ?
		ORDER BY c.timestamp ASC
		LIMIT 100
	`, messageID)

	if err != nil {
		return nil, fmt.Errorf("error getting comments: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.Id, &comment.Author, &comment.Text); err != nil {
			return nil, fmt.Errorf("error scanning comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// getMessageByID retrieve a message by its ID
func (db *appdbimpl) getMessageByID(messageID int64) (*models.Message, error) {
	var msg models.Message
	var text sql.NullString
	var photoBytes []byte
	var timestamp time.Time
	var senderUsername string

	err := db.c.QueryRow(`
		SELECT m.id, u.username, m.type, m.text, m.photo, m.timestamp
		FROM messages m
		INNER JOIN users u ON m.sender_id = u.id
		WHERE m.id = ?
	`, messageID).Scan(&msg.Id, &senderUsername, &msg.Type, &text, &photoBytes, &timestamp)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting message: %w", err)
	}

	msg.Sender = senderUsername
	msg.Timestamp = timestamp

	// Set text if present
	if text.Valid {
		msg.Text = &text.String
	}

	// Convert photo BLOB to base64 if present
	if len(photoBytes) > 0 {
		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		pic := photoBase64
		msg.Photo = &pic
	}

	// Get comment count
	var commentCount int
	err = db.c.QueryRow("SELECT COUNT(*) FROM comments WHERE message_id = ?", messageID).Scan(&commentCount)
	if err != nil {
		return nil, fmt.Errorf("error counting comments: %w", err)
	}
	msg.CommentsCount = commentCount

	// Get comment authors (up to 3)
	rows, err := db.c.Query(`
		SELECT DISTINCT u.username
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		WHERE c.message_id = ?
		ORDER BY c.timestamp DESC
		LIMIT 3
	`, messageID)

	if err != nil {
		return nil, fmt.Errorf("error getting comment authors: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var authors = []string{}
	for rows.Next() {
		var author string
		if err := rows.Scan(&author); err != nil {
			return nil, fmt.Errorf("error scanning author: %w", err)
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating authors: %w", err)
	}

	msg.CommentsAuthors = authors

	return &msg, nil
}
