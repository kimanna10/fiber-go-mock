package handlers

import (
	"context"
	"strconv"
	"time"

	"fiber-go/internal/errs"
	"fiber-go/internal/models"
	"fiber-go/internal/responses"
	"fiber-go/internal/services"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	svc services.UserService
}

// NewUserHandler внедряет зависимость UserService
func NewUserHandler(svc services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// CreateUser — регистрация нового пользователя
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	var req models.UserRegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.svc.Register(ctx, req)
	if err != nil {
		return err
	}

	return responses.Success(c, 201, resp)
}

// GetUsers — получение списка с пагинацией и фильтром
func (h *UserHandler) GetUsers(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	name := c.Query("name")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.svc.GetUsers(ctx, limit, page, name)
	if err != nil {
		return err
	}

	return responses.Success(c, 200, users)
}

// GetUserById — получение одного пользователя по ID
func (h *UserHandler) GetUserById(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return errs.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.svc.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	return responses.Success(c, 200, user)
}

// UpdateUser — частичное обновление (PATCH)
func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return errs.ErrBadRequest
	}

	// var data map[string]interface{}
	var data models.UserUpdateRequest
	if err := c.Bind().Body(&data); err != nil {
		return errs.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	updatedUser, err := h.svc.UpdateUser(ctx, id, data)
	if err != nil {
		return err
	}

	return responses.Success(c, 200, updatedUser)
}

// DeleteUser — удаление пользователя
func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return errs.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	if err := h.svc.DeleteUser(ctx, id); err != nil {
		return err
	}

	return responses.Success(c, 204, nil)
}
