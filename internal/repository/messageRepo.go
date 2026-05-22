package repository

import (
	"context"
	"database/sql"
	"fiber-go/internal/models"
)

type MessageRepository interface {
	Create(ctx context.Context, msg *models.WSMessage) error
	GetByChatId(ctx context.Context, chatId int64) ([]models.WSMessage, error)
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *models.WSMessage) error {
	query := `INSERT INTO messages (chat_id, sender_id, text) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, msg.ChatID, msg.SenderID, msg.Text).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *messageRepository) GetByChatId(ctx context.Context, chatId int64) ([]models.WSMessage, error) {
	query := `SELECT id, chat_id, sender_id, text, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []models.WSMessage
	for rows.Next() {
		var m models.WSMessage
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, rows.Err()
}
