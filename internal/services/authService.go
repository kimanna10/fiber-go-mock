package services

import (
	"context"
	"database/sql"
	"time"

	"fiber-go/internal/auth"
	"fiber-go/internal/errs"
	"fiber-go/internal/models"
	"fiber-go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// 1. Описываем интерфейс со всеми методами
type AuthService interface {
	Login(ctx context.Context, req models.UserLoginRequest) (*models.LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*models.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	generateFullSession(ctx context.Context, user models.User) (*models.LoginResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	authRepo     repository.AuthRepository
	tokenService auth.TokenService
}

func NewAuthService(uRepo repository.UserRepository, aRepo repository.AuthRepository, tService auth.TokenService) AuthService {
	return &authService{
		userRepo:     uRepo,
		authRepo:     aRepo,
		tokenService: tService,
	}
}

// Login — проверяет данные и создает сессию
func (s *authService) Login(ctx context.Context, req models.UserLoginRequest) (*models.LoginResponse, error) {
	// 1. Ищем пользователя через UserRepository
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUnauthorized // Юзер не найден
		}
		return nil, err
	}

	// 2. Сверяем хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errs.ErrUnauthorized // Пароль неверный
	}

	// 3. Создаем токены и сохраняем сессию
	return s.generateFullSession(ctx, user)
}

// Refresh — выдает новые токены, если старый рефреш-токен валиден
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// 1. Хэшируем полученную строку
	hash := s.tokenService.HashToken(refreshToken)

	// 2. Проверяем в базе через AuthRepository
	storedToken, err := s.authRepo.GetRefreshToken(ctx, hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUnauthorized // Токена нет в базе
		}
		return nil, err
	}

	// 3. Проверяем срок годности
	if time.Now().After(storedToken.ExpiresAt) {
		_ = s.authRepo.DeleteRefreshToken(ctx, hash) // Удаляем протухший
		return nil, errs.ErrUnauthorized
	}

	// 4. Достаем юзера, чтобы перевыпустить Access Token с актуальной ролью
	user, err := s.userRepo.GetUserById(ctx, storedToken.UserID)
	if err != nil {
		return nil, errs.ErrInternal
	}

	// 5. Генерируем новую сессию (Refresh Token Rotation)
	return s.generateFullSession(ctx, user)
}

// Logout — просто удаляет сессию из базы
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash := s.tokenService.HashToken(refreshToken)
	return s.authRepo.DeleteRefreshToken(ctx, hash)
}

// Вспомогательный приватный метод, чтобы не дублировать код генерации
func (s *authService) generateFullSession(ctx context.Context, user models.User) (*models.LoginResponse, error) {
	// Генерируем строки токенов (используем твой пакет internal/auth)
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Сохраняем хэш в базу через AuthRepository (на 7 дней)
	hashedRefresh := s.tokenService.HashToken(refreshToken)
	err = s.authRepo.SaveRefreshToken(ctx, user.ID, hashedRefresh, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}
