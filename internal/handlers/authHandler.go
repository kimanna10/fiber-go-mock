package handlers

import (
	"context"
	"time"

	"fiber-go/internal/errs"
	"fiber-go/internal/models"
	"fiber-go/internal/responses"
	"fiber-go/internal/services"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authSvc services.AuthService
}

func NewAuthHandler(aSvc services.AuthService) *AuthHandler {
	return &AuthHandler{
		authSvc: aSvc,
	}
}

// Login: POST /api/login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req models.UserLoginRequest

	// Читаем тело запроса
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrBadRequest
	}

	// Создаем контекст с таймаутом (bcrypt требует времени)
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Вызываем сервис
	resp, err := h.authSvc.Login(ctx, req)
	if err != nil {
		return err
	}

	return responses.Success(c, 200, resp)
}

// Refresh: POST /api/refresh
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	// Ожидаем структуру { "refresh_token": "..." }
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind().Body(&input); err != nil || input.RefreshToken == "" {
		return errs.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Сервис проверит токен и выдаст новую пару
	resp, err := h.authSvc.Refresh(ctx, input.RefreshToken)
	if err != nil {
		return err
	}

	return responses.Success(c, 200, resp)
}

// Logout: POST /api/logout
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind().Body(&input); err != nil || input.RefreshToken == "" {
		return errs.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	// Просто удаляем сессию
	if err := h.authSvc.Logout(ctx, input.RefreshToken); err != nil {
		return err
	}

	return responses.Success(c, 200, fiber.Map{
		"message": "logged out successfully",
	})
}
