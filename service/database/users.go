package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/val7e/wasaText/service/models"
)

// creates a default photo: 5x5 red square PNG in base64
const defaultPhotoBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="

var defaultPhotoBytes []byte

// init runs at first
func init() {
	var err error
	defaultPhotoBytes, err = base64.StdEncoding.DecodeString(defaultPhotoBase64)
	if err != nil {
		panic("Failed to decode default photo: " + err.Error())
	}
}

// DoLogin handles user's login/registration
func (db *appdbimpl) DoLogin(username models.Username) (*models.User, int64, error) {
	// Checks if the user already exists
	var user models.User
	var userID int64
	var picBytes []byte

	err := db.c.QueryRow(
		"SELECT id, username, pic FROM users WHERE username = ?",
		string(username),
	).Scan(&user.Id, &user.Username, &picBytes)

	if err == nil {
		// User already exists - login
		user.Pic = base64.StdEncoding.EncodeToString(picBytes)
		return &user, int64(user.Id), nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, 0, fmt.Errorf("error checking user existence: %w", err)
	}

	// User doesn't exists - registration with default pic
	result, err := db.c.Exec(
		"INSERT INTO users (username, pic) VALUES (?, ?)",
		string(username),
		defaultPhotoBytes,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating user: %w", err)
	}

	userID, err = result.LastInsertId()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting new user ID: %w", err)
	}

	// Return the profile of the new user
	newUser := models.User{
		Id:       models.Id(userID),
		Username: username,
		Pic:      base64.StdEncoding.EncodeToString(defaultPhotoBytes),
	}

	return &newUser, userID, nil
}

// SearchUsers search for a user by username
func (db *appdbimpl) SearchUsers(query models.Username) ([]models.User, error) {
	searchPattern := "%" + string(query) + "%"

	rows, err := db.c.Query(
		"SELECT id, username, pic FROM users WHERE username LIKE ? ORDER BY username LIMIT 700",
		searchPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var users []models.User
	var picBytes []byte
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Id, &user.Username, &picBytes)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// SetMyUserName updates the user's username
func (db *appdbimpl) SetMyUserName(userID int64, newUsername models.Username) (*models.User, error) {
	// checks if the chosen username it's already in use
	var existingID int64
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", string(newUsername)).Scan(&existingID)

	if err == nil && existingID != userID {
		return nil, fmt.Errorf("username already taken")
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error checking username availability: %w", err)
	}

	// updates the username
	_, err = db.c.Exec(
		"UPDATE users SET username = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(newUsername), userID,
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
		return nil, fmt.Errorf("invalid base64 photo data: %w", err)
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
