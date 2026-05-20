package models

type WSMessage struct {
	Type string `json:"type"`
	To   int    `json:"to,omitempty"`
	From int    `json:"from,omitempty"`
	Body string `json:"body,omitempty"`
}
