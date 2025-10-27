package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/val7e/wasaText/service/models"
)

// GetMyConversations retrieves all conversations for a specific user
func (db *appdbimpl) GetMyConversations(userID int64) ([]models.ConversationSummary, error) {
	query := `
		SELECT DISTINCT
			c.id,
			c.type,
			c.name,
			c.convo_pic,
			cs.last_message_timestamp,
			cs.last_message_preview
		FROM conversations c
		INNER JOIN conversation_participants cp ON c.id = cp.conversation_id
		LEFT JOIN conversation_summaries cs ON c.id = cs.id
		WHERE cp.user_id = ?
		ORDER BY cs.last_message_timestamp DESC
		LIMIT 1000
	`

	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var conversations []models.ConversationSummary
	for rows.Next() {
		var conv models.ConversationSummary
		var lastMsgTimestamp sql.NullString
		var lastMsgPreview sql.NullString

		err := rows.Scan(
			&conv.Id,
			&conv.Type,
			&conv.Name,
			&conv.ConvoPic,
			&lastMsgTimestamp,
			&lastMsgPreview,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		// Get participants
		participants, err := db.getConversationParticipants(conv.Id)
		if err != nil {
			return nil, fmt.Errorf("error getting participants: %w", err)
		}
		conv.Participants = participants

		// Set last message if exists
		if lastMsgTimestamp.Valid && lastMsgPreview.Valid {
			timestamp, _ := time.Parse(time.RFC3339, lastMsgTimestamp.String)
			conv.LastMessage = &models.MessagePreview{
				Timestamp: models.Timestamp(timestamp),
				Preview:   lastMsgPreview.String,
			}
		}

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %w", err)
	}

	return conversations, nil
}

// GetConversation retrieves a specific conversation with messages
func (db *appdbimpl) GetConversation(conversationID models.Id, userID int64) (*models.Conversation, error) {
	// Check if user is participant
	var participantCount int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		conversationID, userID,
	).Scan(&participantCount)

	if err != nil {
		return nil, fmt.Errorf("error checking participation: %w", err)
	}
	if participantCount == 0 {
		return nil, fmt.Errorf("user not participant in conversation")
	}

	// Get conversation details
	var conv models.Conversation
	err = db.c.QueryRow(
		"SELECT id, name, type, convo_pic FROM conversations WHERE id = ?",
		conversationID,
	).Scan(&conv.Id, &conv.Name, &conv.Type, &conv.ConvoPic)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting conversation: %w", err)
	}

	// Get participants
	participants, err := db.getConversationParticipants(conversationID)
	if err != nil {
		return nil, fmt.Errorf("error getting participants: %w", err)
	}
	conv.Participants = participants

	// Get messages
	messages, err := db.getConversationMessages(conversationID)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	conv.Messages = messages

	return &conv, nil
}

// StartConversation creates a new direct conversation
func (db *appdbimpl) StartConversation(senderID int64, recipientUsername models.Username) (*models.Conversation, error) {
	// Get recipient user ID
	var recipientID int64
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", string(recipientUsername)).Scan(&recipientID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("recipient user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error finding recipient: %w", err)
	}

	// Check if conversation already exists
	query := `
		SELECT c.id 
		FROM conversations c
		INNER JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		INNER JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		WHERE c.type = 'user' 
		AND cp1.user_id = ? 
		AND cp2.user_id = ?
	`
	var existingConvID int64
	err = db.c.QueryRow(query, senderID, recipientID).Scan(&existingConvID)
	if err == nil {
		// Conversation exists, return it
		return db.GetConversation(models.Id(existingConvID), senderID)
	}

	// Create new conversation
	result, err := db.c.Exec(
		"INSERT INTO conversations (type, created_at, updated_at) VALUES ('user', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
	)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation: %w", err)
	}

	convID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting conversation ID: %w", err)
	}

	// Add participants
	_, err = db.c.Exec(
		"INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?), (?, ?)",
		convID, senderID, convID, recipientID,
	)
	if err != nil {
		return nil, fmt.Errorf("error adding participants: %w", err)
	}

	return db.GetConversation(models.Id(convID), senderID)
}

// Helper function to get conversation participants
func (db *appdbimpl) getConversationParticipants(conversationID models.Id) ([]models.Username, error) {
	rows, err := db.c.Query(`
		SELECT u.username 
		FROM users u
		INNER JOIN conversation_participants cp ON u.id = cp.user_id
		WHERE cp.conversation_id = ?
	`, conversationID)

	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var participants []models.Username
	for rows.Next() {
		var username models.Username
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		participants = append(participants, username)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return participants, rows.Err()
}

// Helper function to get conversation messages
func (db *appdbimpl) getConversationMessages(conversationID models.Id) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT 
			m.id, m.timestamp, u.username, m.type, m.comments_count, m.text, m.photo
		FROM messages m
		INNER JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC
		LIMIT 1000
	`, conversationID)

	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var text sql.NullString
		var photo sql.NullString

		err := rows.Scan(
			&msg.Id,
			&msg.Timestamp,
			&msg.Sender,
			&msg.Type,
			&msg.CommentsCount,
			&text,
			&photo,
		)
		if err != nil {
			return nil, err
		}

		if text.Valid {
			textStr := text.String
			msg.Text = &textStr
		}
		if photo.Valid {
			pic := models.Pic(photo.String)
			msg.Photo = &pic
		}

		// Get comment authors
		commentAuthors, err := db.getMessageCommentAuthors(msg.Id)
		if err == nil {
			msg.CommentsAuthors = commentAuthors
		}

		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, rows.Err()
}

// Helper to get comment authors for a message
func (db *appdbimpl) getMessageCommentAuthors(messageID models.Id) ([]models.Username, error) {
	rows, err := db.c.Query(`
		SELECT DISTINCT u.username
		FROM comments c
		INNER JOIN users u ON c.author_id = u.id
		WHERE c.message_id = ?
		LIMIT 1000
	`, messageID)

	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var authors []models.Username
	for rows.Next() {
		var username models.Username
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		authors = append(authors, username)
	}

	return authors, rows.Err()
}
