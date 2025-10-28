package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"

	"github.com/val7e/wasaText/service/models"
)

// creates a default photo: 5x5 red square PNG in base64
const defaultPhotoBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="

var (
	defaultPhotoBytes []byte
	// Username validation regex
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// init runs at first
func init() {
	var err error
	defaultPhotoBytes, err = base64.StdEncoding.DecodeString(defaultPhotoBase64)
	if err != nil {
		panic("Failed to decode default photo: " + err.Error())
	}
}

// validateUsername checks if username meets requirements
func validateUsername(username string) error {
	// Check length (3-25 characters)
	if len(username) < 3 || len(username) > 25 {
		return fmt.Errorf("username must be between 3 and 25 characters")
	}

	// Check pattern: alphanumeric, _, and - only
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, _, and -")
	}

	return nil
}

// DoLogin handles user's login/registration
func (db *appdbimpl) DoLogin(username string) (*models.User, bool, error) {
	// Validate username format
	if err := validateUsername(username); err != nil {
		return nil, false, err
	}

	// Checks if the user already exists
	var user models.User
	var picBytes []byte
	err := db.c.QueryRow(
		"SELECT id, username, pic FROM users WHERE username = ?",
		username,
	).Scan(&user.Id, &user.Username, &picBytes)

	if err == nil {
		// User already exists - login
		user.Pic = base64.StdEncoding.EncodeToString(picBytes)
		return &user, false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("error checking user existence: %w", err)
	}

	// User doesn't exist - registration with default pic
	result, err := db.c.Exec(
		"INSERT INTO users (username, pic) VALUES (?, ?)",
		username,
		defaultPhotoBytes,
	)
	if err != nil {
		return nil, false, fmt.Errorf("error creating user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, false, fmt.Errorf("error getting new user ID: %w", err)
	}

	// Return the profile of the new user
	newUser := models.User{
		Id:       userID,
		Username: username,
		Pic:      base64.StdEncoding.EncodeToString(defaultPhotoBytes),
	}

	return &newUser, true, nil
}

// SearchUser searches for users by username pattern
func (db *appdbimpl) SearchUser(query string) ([]models.User, error) {
	searchPattern := "%" + query + "%"
	rows, err := db.c.Query(
		"SELECT id, username, pic FROM users WHERE username LIKE ? ORDER BY username LIMIT 700",
		searchPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []models.User
	for rows.Next() {
		var user models.User
		var picBytes []byte
		err := rows.Scan(&user.Id, &user.Username, &picBytes)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		user.Pic = base64.StdEncoding.EncodeToString(picBytes)
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// SetMyUserName updates the user's username
func (db *appdbimpl) SetMyUserName(userID int64, newUsername string) (*models.User, error) {
	// Validate new username format
	if err := validateUsername(newUsername); err != nil {
		return nil, err
	}

	// checks if the chosen username is already in use
	var existingID int64
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", newUsername).Scan(&existingID)
	if err == nil && existingID != userID {
		return nil, fmt.Errorf("username already taken")
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error checking username availability: %w", err)
	}

	// updates the username
	_, err = db.c.Exec(
		"UPDATE users SET username = ? WHERE id = ?",
		newUsername, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating username: %w", err)
	}

	// returns the updated user
	return db.GetUserByID(userID)
}

// SetMyPhoto updates the user's profile picture
func (db *appdbimpl) SetMyPhoto(userID int64, newPicBase64 string) (*models.User, error) {
	// Decode base64 string to binary data
	picBytes, err := base64.StdEncoding.DecodeString(newPicBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 photo data")
	}

	// Update the photo in database as BLOB
	_, err = db.c.Exec(
		"UPDATE users SET pic = ? WHERE id = ?",
		picBytes,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating user pic: %w", err)
	}

	// Returns the updated user
	return db.GetUserByID(userID)
}

// GetUserByID retrieves a user by ID
func (db *appdbimpl) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	var picBytes []byte
	err := db.c.QueryRow(
		"SELECT id, username, pic FROM users WHERE id = ?",
		userID,
	).Scan(&user.Id, &user.Username, &picBytes)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	user.Pic = base64.StdEncoding.EncodeToString(picBytes)
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (db *appdbimpl) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	var picBytes []byte
	err := db.c.QueryRow(
		"SELECT id, username, pic FROM users WHERE username = ?",
		username,
	).Scan(&user.Id, &user.Username, &picBytes)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	user.Pic = base64.StdEncoding.EncodeToString(picBytes)
	return &user, nil
}
