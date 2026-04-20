package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

//
// CONFIG
//

type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
}

// Готовые пресеты
var (
	StrictLimit = RateLimitConfig{
		Max:        5,
		Expiration: time.Minute,
	}

	NormalLimit = RateLimitConfig{
		Max:        100,
		Expiration: time.Minute,
	}

	RelaxedLimit = RateLimitConfig{
		Max:        1000,
		Expiration: time.Minute,
	}
)

//
// KEY GENERATORS
//

// По IP (для login, public)
func KeyByIP(c fiber.Ctx) string {
	return c.IP()
}

// По user_id (если есть), иначе IP
func KeyByUserOrIP(c fiber.Ctx) string {
	userID := c.Locals("user_id")
	if userID != nil {
		return fmt.Sprintf("user:%v", userID)
	}
	return c.IP()
}

//
// MIDDLEWARE
//

func RateLimit(cfg RateLimitConfig, keyGen func(fiber.Ctx) string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,

		KeyGenerator: keyGen,

		LimitReached: func(c fiber.Ctx) error {
			// через сколько можно снова
			c.Set("Retry-After", fmt.Sprintf("%d", int(cfg.Expiration.Seconds())))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many requests",
				"retry_after": cfg.Expiration.Seconds(),
			})
		},
	})
}
