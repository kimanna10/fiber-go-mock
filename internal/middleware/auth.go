package middleware

import (
	"fiber-go/internal/auth"
	"fiber-go/internal/errs"
	"fiber-go/internal/patterns"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Константы для ключей, чтобы не ошибиться в буквах
const (
	UserIDKey = "user_id"
	RoleKey   = "role"
)

// JWTMiddleware проверяет наличие и валидность токена
// func JWTMiddleware(tokenService auth.TokenService) fiber.Handler {
// 	return func(c fiber.Ctx) error {
// 		authHeader := c.Get("Authorization")
// 		if authHeader == "" {
// 			return errs.ErrUnauthorized
// 		}
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			return errs.ErrUnauthorized
// 		}

// 		claims, err := tokenService.ParseToken(parts[1])
// 		if err != nil {
// 			return errs.ErrInvalidToken
// 		}

// 		// Используем константы вместо строк
// 		c.Locals(UserIDKey, claims.UserID)
// 		c.Locals(RoleKey, claims.Role)

//			return c.Next()
//		}
//	}
//
// JWTMiddleware проверяет наличие и валидность токена (теперь и в Query для WebSocket)
func JWTMiddleware(tokenService auth.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		var tokenStr string

		if authHeader != "" {
			// Если токен идет в хедерах (стандартный HTTP)
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return errs.ErrUnauthorized
			}
			tokenStr = parts[1]
		} else {
			// [+] Если в хедерах пусто, проверяем query-параметр ?token=... (для WebSocket)
			tokenStr = c.Query("token")
		}

		// Если токена нет вообще нигде — отдаем 401
		if tokenStr == "" {
			return errs.ErrUnauthorized
		}

		// Парсим токен (неважно, откуда мы его достали)
		claims, err := tokenService.ParseToken(tokenStr)
		if err != nil {
			return errs.ErrInvalidToken
		}

		// Используем константы вместо строк
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(RoleKey, claims.Role)

		return c.Next()
	}
}

// RequireRole проверяет, входит ли роль юзера в список разрешенных
func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if len(roles) == 0 {
			return errs.ErrForbidden
		}
		userRole, ok := c.Locals(RoleKey).(string)
		fmt.Println(userRole)
		if !ok {
			return errs.ErrUnauthorized
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}
		return errs.ErrForbidden
	}
}

// CanAccessUser проверяет, является ли юзер владельцем данных или админом
// func CanAccessUser() fiber.Handler {
// 	return func(c fiber.Ctx) error {
// 		// Берем ID из контекста (кто делает запрос)
// 		userID, ok := c.Locals(UserIDKey).(int)
// 		if !ok {
// 			return errs.ErrUnauthorized
// 		}

// 		role, _ := c.Locals(RoleKey).(string)

// 		// Берем ID из параметров URL (над кем делаем запрос)
// 		targetID, err := strconv.Atoi(c.Params("id"))
// 		if err != nil {
// 			return errs.ErrBadRequest
// 		}

// 		// Если я не тот, за кого себя выдаю И я не админ
// 		if userID != targetID && role != "admin" {
// 			return errs.ErrForbidden
// 		}

// 		return c.Next()
// 	}
// }

// CanAccessUser проверяет, является ли юзер владельцем данных или админом c паттерном Strategy
func CanAccessUser(strategy patterns.AccessStrategy) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Берем ID из контекста (кто делает запрос)
		userID, ok := c.Locals(UserIDKey).(int)
		if !ok {
			return errs.ErrUnauthorized
		}
		role, _ := c.Locals(RoleKey).(string)

		// Берем ID из параметров URL (над кем делаем запрос)
		targetID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return errs.ErrBadRequest
		}

		ctx := patterns.AccessContext{
			UserId:   userID,
			TargetId: targetID,
			Role:     patterns.Role(role),
		}

		// Если я не тот, за кого себя выдаю И я не админ
		access := strategy.CanAccess(ctx)
		if !access {
			return errs.ErrForbidden
		}
		return c.Next()
	}
}
