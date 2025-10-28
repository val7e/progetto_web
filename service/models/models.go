package models

import "time"

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Pic      string `json:"pic"`
}

type Group struct {
	Id         int64    `json:"id"`
	Name       string   `json:"name"`
	Members    []string `json:"members"`
	GroupPhoto *string  `json:"group_photo,omitempty"`
}

type Conversation struct {
	Id           int64           `json:"id"`
	Name         *string         `json:"name,omitempty"`
	Type         string          `json:"type"`
	Participants []string        `json:"participants"`
	ConvoPic     *string         `json:"convo_pic,omitempty"`
	LastMessage  *MessagePreview `json:"last_message,omitempty"`
	Messages     []Message       `json:"messages"`
}

type MessagePreview struct {
	Timestamp time.Time `json:"timestamp"`
	Preview   string    `json:"preview"`
}
type Message struct {
	Id              int64     `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Sender          string    `json:"sender"`
	Type            string    `json:"type"`
	CommentsCount   int       `json:"comments_count"`
	CommentsAuthors []string  `json:"comments_authors"`

	Text  *string `json:"text,omitempty"`
	Photo *string `json:"photo,omitempty"`
}

type NewMessage struct {
	Sender string  `json:"sender"`
	Type   string  `json:"type"`
	Text   *string `json:"text,omitempty"`
	Photo  *string `json:"photo,omitempty"`
}

type ConversationSummary struct {
	Id           int64           `json:"id"`
	Type         string          `json:"type"`
	Name         *string         `json:"name,omitempty"`
	Participants []string        `json:"participants"`
	ConvoPic     *string         `json:"convo_pic,omitempty"`
	LastMessage  *MessagePreview `json:"last_message,omitempty"`
}

type Comment struct {
	Id     int64  `json:"id"`
	Author string `json:"username"`
	Text   string `json:"text"`
}

type NewComment struct {
	Text string `json:"text"`
}
