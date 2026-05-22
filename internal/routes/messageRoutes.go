package routes

import (
	"fiber-go/internal/auth"
	"fiber-go/internal/handlers"
	"fiber-go/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// MessageRoutes регистрирует маршруты для работы с сообщениями и историей чатов
func MessageRoutes(app *fiber.App, h *handlers.MessageHandler, jwtService auth.TokenService) {

	// Создаем защищенную группу для сообщений
	messages := app.Group(
		"/messages",
		middleware.JWTMiddleware(jwtService),
		// Защита от спама запросами истории чата
		middleware.RateLimit(middleware.NormalLimit, middleware.KeyByUserOrIP),
	)

	// GET /messages?chat_id=...
	// Возвращает историю сообщений. Доступно любому авторизованному пользователю
	messages.Get("/", h.GetChatHistory)
}
