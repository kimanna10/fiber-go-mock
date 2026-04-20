package routes

import (
	"fiber-go/internal/handlers"
	"fiber-go/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// Передаем сюда инициализированный userHandler
func UserRoutes(app *fiber.App, h *handlers.UserHandler) {
	// Группа роутов, защищенная JWT
	users := app.Group("/users", middleware.JWTMiddleware, middleware.RateLimit(middleware.NormalLimit, middleware.KeyByUserOrIP))

	// Теперь используем методы структуры через переменную h
	users.Get("/", middleware.RequireRole("admin"), h.GetUsers)
	users.Get("/:id", middleware.CanAccessUser(), h.GetUserById)

	// Мы решили объединить логику в один Patch или оставить UpdateUser
	users.Patch("/:id", middleware.CanAccessUser(), h.UpdateUser)

	users.Delete("/:id", middleware.CanAccessUser(), h.DeleteUser)
}
