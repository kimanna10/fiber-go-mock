package models

import (
	"encoding/json"
)

type EventType string

const (
	EventMessage      EventType = "message"
	EventTyping       EventType = "typing"
	EventNotification EventType = "notification"
)

// Наш главный конверт для сети
type WSEvent struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Временное событие, живущее только в сокетах
type TypingEvent struct {
	ChatID int64 `json:"chat_id"`
	UserID int   `json:"user_id"`
}
