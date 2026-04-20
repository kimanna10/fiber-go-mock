package app

import (
	"database/sql"
	"fiber-go/internal/handlers"
	"fiber-go/internal/middleware"
	"fiber-go/internal/repository"
	"fiber-go/internal/responses"
	"fiber-go/internal/routes"
	"fiber-go/internal/services"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// Setup связывает все слои приложения и возвращает настроенный Fiber App
func Setup(db *sql.DB, log *slog.Logger) *fiber.App {

	// 1. Инициализация слоев (Dependency Injection)
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)

	userSvc := services.NewUserService(userRepo)
	authSvc := services.NewAuthService(userRepo, authRepo)

	userHdl := handlers.NewUserHandler(userSvc)
	authHdl := handlers.NewAuthHandler(authSvc)

	// 2. Настройка Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: responses.Error,
	})

	// 3. Глобальные Middleware
	app.Use(middleware.Logger(log))

	// 4. Регистрация роутов
	// Публичные роуты (без JWT)
	api := app.Group("/api")
	api.Post("/register", userHdl.CreateUser)

	api.Post("/login", middleware.RateLimit(middleware.StrictLimit, middleware.KeyByIP), authHdl.Login)
	api.Post("/refresh", authHdl.Refresh)
	api.Post("/logout", authHdl.Logout)

	// Приватные роуты (внутри пакета routes)
	routes.UserRoutes(app, userHdl)

	return app
}
