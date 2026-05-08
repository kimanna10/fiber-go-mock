package routes

import (
	"fiber-go/internal/auth"
	"fiber-go/internal/handlers"
	"fiber-go/internal/middleware"
	"fiber-go/internal/patterns"

	"github.com/gofiber/fiber/v3"
)

// Передаем сюда инициализированный userHandler
func UserRoutes(app *fiber.App, h *handlers.UserHandler) {

	jwtService := auth.NewJWTService()

	// Группа роутов, защищенная JWT и с ограничением по количеству запросов
	users := app.Group(
		"/users",
		middleware.JWTMiddleware(jwtService),
		middleware.RateLimit(middleware.NormalLimit, middleware.KeyByUserOrIP),
	)

	accessUser := middleware.CanAccessUser(patterns.OwnerOrAdmin())

	// Теперь используем методы структуры через переменную h
	users.Get("/",
		middleware.RequireRole("admin"),
		h.GetUsers,
	)

	users.Get("/:id",
		accessUser,
		h.GetUserById,
	)
	// Мы решили объединить логику в один Patch или оставить UpdateUser
	users.Patch("/:id", accessUser, h.UpdateUser)
	users.Delete("/:id", accessUser, h.DeleteUser)
}
