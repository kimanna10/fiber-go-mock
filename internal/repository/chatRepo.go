package repository

import (
	"context"
	"database/sql"
	"fiber-go/internal/models"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) error
	AddMember(ctx context.Context, req *models.AddMemberRequest) error
	GetChatInfo(ctx context.Context, chatId int64) (*models.ChatInfo, error)
}

type chatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *models.Chat) error {
	query := `INSERT INTO chats (type, title) VALUES ($1, $2) RETURNING id, created_t`

	var sqlTitle sql.NullString
	if chat.Title != "" {
		sqlTitle = sql.NullString{String: chat.Title, Valid: true}
	}
	return r.db.QueryRowContext(ctx, query, chat.Type, sqlTitle).Scan(&chat.ID, &chat.CreatedAt)
}

func (r *chatRepository) AddMember(ctx context.Context, req *models.AddMemberRequest) error {
	query := `INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2) RETURNING id ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, req.ChatId, req.UserId)
	return err
}

func (r *chatRepository) GetChatInfo(ctx context.Context, chatId int64) (*models.ChatInfo, error) {
	var chatType models.ChatType
	query := `SELECT type FROM chats WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, chatId).Scan(&chatType)
	if err != nil {
		return nil, err
	}
	info := &models.ChatInfo{Type: chatType}
	if chatType == models.ChatBroadcast {
		return info, nil
	}

	queryGetMembers := `SELECT user_id FROM chat_members WHERE chat_id = $1`
	rows, err := r.db.QueryContext(ctx, queryGetMembers, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid int
		_ = rows.Scan(&uid)
		info.Members = append(info.Members, uid)
	}
	return info, nil
}
