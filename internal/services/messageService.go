package services

import (
	"context"
	"fiber-go/internal/errs"
	"fiber-go/internal/models"
	"fiber-go/internal/repository"
)

type MessageService interface {
	SaveMessage(ctx context.Context, msg *models.WSMessage) error
	GetChatHistory(ctx context.Context, chatID int64) ([]models.WSMessage, error)
}

type messageService struct {
	repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

func (s *messageService) SaveMessage(ctx context.Context, msg *models.WSMessage) error {
	if msg.Text == "" || msg.ChatID == 0 || msg.SenderID == 0 {
		return errs.ErrBadRequest
	}
	return s.repo.Create(ctx, msg)
}

func (s *messageService) GetChatHistory(ctx context.Context, chatID int64) ([]models.WSMessage, error) {
	if chatID == 0 {
		return nil, errs.ErrBadRequest
	}

	messages, err := s.repo.GetByChatId(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// Переворачиваем с DESC на ASC, чтобы история шла сверху вниз
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
