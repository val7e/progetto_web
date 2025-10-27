package models

import "time"

type Id int64
type Username string
type Name string
type Pic string
type Timestamp time.Time
type ConvoType string

type LoginRequest struct {
	Username Username `json:"username"`
}

type LoginResponse struct {
	Id   int64 `json:"id"`
	User User  `json:"user"`
}

type User struct {
	Id       Id       `json:"id"`
	Username Username `json:"username"`
	Pic      string   `json:"pic"`
}

type Group struct {
	Id         Id         `json:"id"`
	Name       Name       `json:"name"`
	Members    []Username `json:"members"`
	GroupPhoto *Pic       `json:"group_photo,omitempty"`
}

type Conversation struct {
	Id           Id              `json:"id"`
	Name         *Name           `json:"name,omitempty"`
	Type         ConvoType       `json:"type"`
	Participants []Username      `json:"participants"`
	ConvoPic     *Pic            `json:"convo_pic,omitempty"`
	LastMessage  *MessagePreview `json:"last_message,omitempty"`
	Messages     []Message       `json:"messages"`
}

type MessagePreview struct {
	Timestamp Timestamp `json:"timestamp"`
	Preview   string    `json:"preview"`
}
type Message struct {
	Id              Id         `json:"id"`
	Timestamp       Timestamp  `json:"timestamp"`
	Sender          Username   `json:"sender"`
	Type            string     `json:"type"`
	CommentsCount   int        `json:"comments_count"`
	CommentsAuthors []Username `json:"comments_authors"`

	Text  *string `json:"text,omitempty"`
	Photo *Pic    `json:"photo,omitempty"`
}

type NewMessage struct {
	Type  string  `json:"type"`
	Text  *string `json:"text,omitempty"`
	Photo *Pic    `json:"photo,omitempty"`
}

type ConversationSummary struct {
	Id           Id              `json:"id"`
	Type         ConvoType       `json:"type"`
	Name         *string         `json:"name,omitempty"`
	Participants []Username      `json:"participants"`
	ConvoPic     *Pic            `json:"convo_pic,omitempty"`
	LastMessage  *MessagePreview `json:"last_message,omitempty"`
}

type Comment struct {
	Id     Id       `json:"id"`
	Author Username `json:"username"`
	Text   string   `json:"text"`
}

type NewComment struct {
	Text string `json:"text"`
}
