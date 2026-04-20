package responses

import (
	"fiber-go/internal/errs"

	"github.com/gofiber/fiber/v3"
)

func Success(c fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"data":  data,
		"error": nil,
	})
}

func Error(c fiber.Ctx, err error) error {
	// 1. свои ошибки
	if appErr, ok := err.(*errs.AppError); ok {
		return c.Status(appErr.StatusCode).JSON(fiber.Map{
			"data":  nil,
			"error": appErr.Message,
		})
	}

	// 2. fiber (fallback)
	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"data":  nil,
			"error": fiberErr.Message,
		})
	}

	return c.Status(500).JSON(fiber.Map{
		"data":  nil,
		"error": "Internal server error",
	})
}
