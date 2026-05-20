package main

import (
	"context"
	"encoding/json"
	"fiber-go/internal/config"
	"fiber-go/internal/logger"
	"fiber-go/internal/services"

	"github.com/hibiken/asynq"
)

func main() {
	// 1. Окружение
	cfg := config.Load()

	// 2. Инициализируем логгер первым
	log := logger.New()

	emailSrv := services.NewEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.From,
		cfg.SMTP.User,
		cfg.SMTP.Pass,
		log,
	)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Host + ":" + cfg.Redis.Port},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()

	mux.HandleFunc("email:welcome", func(ctx context.Context, t *asynq.Task) error {
		var p struct {
			Email string
			Name  string
		}
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		log.Info("Worker: Отправка сообщения", "to", p.Email)
		return emailSrv.SendWelcomeEmail(p.Email, p.Name)
	})

	log.Info("Worker работает")
	if err := srv.Run(mux); err != nil {
		log.Error("Worker ошибка", "error", err)
	}
}
