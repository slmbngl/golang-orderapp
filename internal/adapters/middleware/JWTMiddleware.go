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

		// Get user_id and role from token and save to context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, exists := claims["user_id"]; exists {
				c.Locals("user_id", int(userID.(float64)))
			}
			if role, exists := claims["role"]; exists {
				c.Locals("role", role.(string))
			}
		}

		return c.Next()
	}
}

// RoleMiddleware creates a middleware that checks for specific roles
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No role information found in token",
			})
		}

		role := userRole.(string)

		// Check if user's role is in the allowed roles
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// AdminMiddleware is a convenience function for admin-only endpoints
func AdminMiddleware() fiber.Handler {
	return RoleMiddleware("admin")
}
