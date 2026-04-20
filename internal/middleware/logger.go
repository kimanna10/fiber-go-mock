package middleware

import (
	"fiber-go/internal/errs"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// Выполняем запрос
		err := c.Next()

		// Если произошла ошибка, Fiber может не успеть обновить статус в ответе
		// Берем статус из контекста или дефолтный 200
		status := c.Response().StatusCode()
		if err != nil && status == 200 {
			if e, ok := err.(*errs.AppError); ok {
				status = e.StatusCode
			} else {
				status = 500
			}
		}

		duration := time.Since(start)

		// Подготавливаем базовые поля
		fields := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration", duration,
		}

		// Если была ошибка, добавляем её в логи
		if err != nil {
			fields = append(fields, "error", err.Error())
		}

		// Логируем в зависимости от статуса
		switch {
		case status >= 500:
			log.Error("SERVER_ERROR", fields...)
		case status >= 400:
			log.Warn("CLIENT_ERROR", fields...)
		default:
			// Для обычных запросов логируем кратко
			log.Info("HTTP_REQUEST", fields...)
		}

		return err
	}
}

// func Logger(log *slog.Logger) fiber.Handler {
// 	return func(c fiber.Ctx) error {
// 		start := time.Now()
// 		err := c.Next() // Получаем ошибку из responses.Error

// 		status := c.Response().StatusCode()
// 		// Страховка: если статус не успел обновиться, берем его из ошибки
// 		if err != nil && status == 200 {
// 			if e, ok := err.(*errs.AppError); ok {
// 				status = e.StatusCode
// 			} else {
// 				status = 500
// 			}
// 		}

// 		duration := time.Since(start)
// 		fields := []any{
// 			"method", c.Method(),
// 			"path", c.Path(),
// 			"status", status,
// 			"duration", duration,
// 		}

// 		if err != nil {
// 			fields = append(fields, "error", err.Error())
// 			log.Error("HTTP_ERROR", fields...)
// 		} else {
// 			log.Info("HTTP_SUCCESS", fields...)
// 		}

// 		return err
// 	}
// }
