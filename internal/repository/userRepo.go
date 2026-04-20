package repository

import (
	"context"
	"database/sql"
	"fiber-go/internal/models"
	"fmt"
)

// 1. Описываем интерфейс со всеми методами
type UserRepository interface {
	GetUsers(ctx context.Context, limit, offset int, name string) ([]models.User, error)
	GetUserById(ctx context.Context, id int) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
}

// UserRepository управляет операциями с таблицей users
type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый экземпляр репозитория
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetUsers возвращает список пользователей с фильтрацией и пагинацией
func (r *userRepository) GetUsers(ctx context.Context, limit, offset int, name string) ([]models.User, error) {
	query := `
        SELECT id, name, age, email, role
        FROM users
        WHERE 1=1
    `
	args := []interface{}{}
	argID := 1

	if name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argID)
		args = append(args, "%"+name+"%")
		argID++
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetUserById находит пользователя по его ID
func (r *userRepository) GetUserById(ctx context.Context, id int) (models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, age, email, role FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.Role)

	return user, err
}

// GetUserByEmail находит пользователя по email (нужно для авторизации)
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, age, email, password, role FROM users WHERE email = $1",
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Age,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	return user, err
}

// Create создает нового пользователя и возвращает его с заполненным ID
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (name, age, email, password, role) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Name,
		user.Age,
		user.Email,
		user.Password,
		user.Role,
	).Scan(&user.ID)
	return err
}

// Update обновляет данные пользователя. Возвращает sql.ErrNoRows, если id не найден
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE users SET name = $1, age = $2, email = $3 WHERE id = $4",
		user.Name,
		user.Age,
		user.Email,
		user.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM users WHERE id = $1",
		id,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
