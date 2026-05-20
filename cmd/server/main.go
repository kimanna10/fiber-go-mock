package main

import (
	"context"
	"fiber-go/internal/app"
	"fiber-go/internal/config"
	"fiber-go/internal/database"
	"fiber-go/internal/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// 1. Окружение
	cfg := config.Load()

	// 2. Инициализируем логгер первым
	log := logger.New()

	// 3. Подключаем базу (теперь передаем вложенную структуру)
	db := database.Connect(cfg.DB)
	defer db.Close() // Закрываем при выключении

	// 3. Собираем всё приложение через App Setup
	server := app.Setup(db, log, cfg)

	// канал для сигнала
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Info("сервер запущен", "port", cfg.Port)
		if err := server.Listen(":" + cfg.Port); err != nil {
			log.Error("сервер ошибка", "error", err)
		}
	}()

	<-quit
	log.Info("сигнал остановки получен, начинаем graceful shutdown...")

	// Gracefull Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.ShutdownWithContext(ctx); err != nil {
		log.Error("ошибка graceful shutdown", "error", err)
	} else {
		log.Info("сервер остановлен")
	}
}
