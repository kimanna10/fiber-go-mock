package repository

import (
	"context"
	"database/sql"
	"time"
)

// 1. Описываем интерфейс со всеми методами
type AuthRepository interface {
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenEntity, error)
	SaveRefreshToken(ctx context.Context, userID int, tokenHash string, duration time.Duration) error
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type RefreshTokenEntity struct {
	UserID    int
	ExpiresAt time.Time
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenEntity, error) {
	var rt RefreshTokenEntity
	// Хорошая практика — выносить запросы в константы или переменные для читаемости
	query := `SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(&rt.UserID, &rt.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *authRepository) SaveRefreshToken(ctx context.Context, userID int, tokenHash string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)

	// Использование ON CONFLICT позволяет обновить существующий токен для юзера,
	// вместо того чтобы плодить дубликаты или выдавать ошибку.
	query := `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at) 
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) 
        DO UPDATE SET token_hash = EXCLUDED.token_hash, expires_at = EXCLUDED.expires_at
    `

	_, err := r.db.ExecContext(ctx, query, userID, tokenHash, expiresAt)
	return err
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}
