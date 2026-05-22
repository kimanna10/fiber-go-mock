package handlers

import (
	"context"
	"strconv"
	"time"

	"fiber-go/internal/errs"
	"fiber-go/internal/responses"
	"fiber-go/internal/services"

	"github.com/gofiber/fiber/v3"
)

type MessageHandler struct {
	svc services.MessageService
}

// NewMessageHandler внедряет зависимость MessageService
func NewMessageHandler(svc services.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

// GetChatHistory — получение истории сообщений конкретного чата
func (h *MessageHandler) GetChatHistory(c fiber.Ctx) error {
	chatIdParam := c.Query("chat_id")
	if chatIdParam == "" {
		return errs.ErrBadRequest
	}

	chatId, err := strconv.ParseInt(chatIdParam, 10, 64)
	if err != nil {
		return errs.ErrBadRequest
	}

	// Устанавливаем таймаут на чтение тяжелой истории из базы
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	messages, err := h.svc.GetChatHistory(ctx, chatId)
	if err != nil {
		return err
	}

	// Отдаем статус 200 и массив сообщений в твоем стандартном формате
	return responses.Success(c, 200, messages)
}
