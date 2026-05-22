package models

import "time"

type ChatType string

const (
	ChatPrivate   ChatType = "private"
	ChatGroup     ChatType = "group"
	ChatBroadcast ChatType = "broadcast"
)

type Chat struct {
	ID        int64     `json:"id"`
	Type      ChatType  `json:"type"`
	Title     string    `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Данные для добавления участников конференции (Request DTO)
type AddMemberRequest struct {
	ChatId string `json:"chat_id"`
	UserId string `json:"user_id"`
}

type ChatInfo struct {
	Type    ChatType
	Members []int
}
