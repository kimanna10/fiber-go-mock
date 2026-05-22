package app

import (
	"database/sql"
	"fiber-go/internal/auth"
	"fiber-go/internal/cache"
	"fiber-go/internal/config"
	"fiber-go/internal/handlers"
	"fiber-go/internal/middleware"
	"fiber-go/internal/repository"
	"fiber-go/internal/responses"
	"fiber-go/internal/routes"
	"fiber-go/internal/services"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/hibiken/asynq"
)

// Setup связывает все слои приложения и возвращает настроенный Fiber App
func Setup(db *sql.DB, log *slog.Logger, cfg *config.Config) *fiber.App {

	keys := cache.NewKeyBuilder("fibergo")
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port

	// 1. Инициализация слоев (Dependency Injection)
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)

	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	redisCache := cache.NewRedisCache(redisAddr)

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	jwtService := auth.NewJWTService(cfg.JWTSecret)

	userSvc := services.NewUserService(userRepo, redisCache, keys, asynqClient)
	authSvc := services.NewAuthService(userRepo, authRepo, jwtService)
	msgSvc := services.NewMessageService(messageRepo)

	userHdl := handlers.NewUserHandler(userSvc)
	authHdl := handlers.NewAuthHandler(authSvc)
	msgHdl := handlers.NewMessageHandler(msgSvc)

	// [+] Инициализация WebSocket слоев
	wsHub := services.NewWSHub(redisCache, keys, messageRepo, chatRepo) // Создаем хаб (сервис)
	wsHdl := handlers.NewWSHandler(wsHub)                               // Создаем хендлер

	// 2. Настройка Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: responses.Error,
	})
	// [+][Глобальный CORS — ставить строго ПЕРЕД роутами и логгером]
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Позволяет делать запросы с любого локального файла/домена
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	// 3. Глобальные Middleware
	app.Use(middleware.Logger(log))

	// 4. Регистрация роутов
	// Публичные роуты (без JWT)
	api := app.Group("/api")
	api.Post("/register", userHdl.CreateUser)

	api.Post("/login", middleware.RateLimit(middleware.StrictLimit, middleware.KeyByIP), authHdl.Login)
	api.Post("/refresh", authHdl.Refresh)
	api.Post("/logout", authHdl.Logout)

	// [+] Маршрут для WebSocket подключений
	// Сначала идет твой JWT мидлвар, чтобы вытащить userID, затем хендлер вебсокета
	app.Get("/ws", middleware.JWTMiddleware(jwtService), wsHdl.HandleWS)

	// Приватные роуты (внутри пакета routes)
	routes.UserRoutes(app, userHdl, jwtService)
	routes.MessageRoutes(app, msgHdl, jwtService)

	return app
}
