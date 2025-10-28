/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/val7e/wasaText/service/models"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error
	Ping() error

	// User operations defined in users.go
	DoLogin(username string) (*models.User, bool, error)
	SearchUser(query string) ([]models.User, error)
	SetMyUserName(userID int64, newUsername string) (*models.User, error)
	SetMyPhoto(userID int64, newPic string) (*models.User, error)
	GetUserByID(userID int64) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)

	// Conversation operations defined in conversations.go
	GetMyConversations(userID int64) ([]models.ConversationSummary, error)
	GetConversation(conversationID int64, userID int64) (*models.Conversation, error)
	StartConversation(senderID int64, recipientUsername string) (*models.Conversation, error)

	// Group operations defined in groups.go
	CreateGroup(creatorID int64, name string) (*models.Group, error)
	GetGroup(groupID int64) (*models.Group, error)
	SetGroupName(groupID int64, name string) (*models.Group, error)
	SetGroupPhoto(groupID int64, photo string) (*models.Group, error)
	AddToGroup(groupID int64, memberUsernames []string) (*models.Group, error)
	LeaveGroup(groupID int64, userID int64) error

	// Message operations defined in messages.go
	SendMessage(conversationID int64, senderID int64, message models.NewMessage) (*models.Message, error)
	ForwardMessage(messageID, recipientConversationID int64, authorID int64) (*models.Message, error)
	DeleteMessage(messageID, conversationID int64, userID int64) error

	// Comment operations defined in messages.go
	CommentMessage(messageID, conversationID int64, authorID int64, comment models.NewComment) (*models.Comment, error)
	UncommentMessage(messageID, conversationID int64, userID int64) error
	GetComments(messageID int64) ([]models.Comment, error)
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Create all necessary tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

// Database tables
func createTables(db *sql.DB) error {
	tables := []string{
		// users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			pic BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// conversations table
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			type TEXT NOT NULL CHECK (type IN ('user', 'group')),
			convo_pic BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// conversation_participants table
		`CREATE TABLE IF NOT EXISTS conversation_participants (
			conversation_id INTEGER,
			user_id INTEGER,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// groups table
		`CREATE TABLE IF NOT EXISTS groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			group_photo BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// group_members table
		`CREATE TABLE IF NOT EXISTS group_members (
			group_id INTEGER,
			user_id INTEGER,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, user_id),
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// messages table
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('text', 'photo')),
			text TEXT,
			photo BLOB,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			CHECK ((type = 'text' AND text IS NOT NULL) OR (type = 'photo' AND photo IS NOT NULL))
		);`,

		// comments table
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			text TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// Index for faster comment lookups
		`CREATE INDEX IF NOT EXISTS idx_comments_message ON comments(message_id);`,

		// Index for faster message lookups
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, timestamp DESC);`,
	}

	// Execute creation queries
	for i, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error creating table/index %d: %w", i+1, err)
		}
	}

	return nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
