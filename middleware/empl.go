package middleware

import (
	"strings"

	"toxkurd.com/bagpost/util/empl"

	"github.com/gofiber/fiber/v3"
)

func EmplAuthMiddleware(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization format"})
	}

	claims, err := empl.ParseToken(parts[1]) // warehouse-specific JWT utils
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	// Check if user is active (optional)
	if isActive, ok := claims["is_active"].(bool); ok && !isActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Account is inactive"})
	}

	c.Locals("id", claims["id"])
	c.Locals("email", claims["email"])

	return c.Next()
}
