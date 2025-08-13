package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/slmbngl/OrderAplication/internal/service"
)

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Get("Authorization")

		// Check if Authorization header is present and starts with "Bearer "
		if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required (Bearer token)",
			})
		}

		// Remove Bearer prefix
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		token, err := service.ParseJWT(tokenStr)
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Get user_id from token and save to context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, exists := claims["user_id"]; exists {
				c.Locals("user_id", int(userID.(float64)))
			}
		}

		return c.Next()
	}
}
