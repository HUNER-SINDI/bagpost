package middleware

import (
	"strings"

	utils "toxkurd.com/bagpost/util/admin"

	"github.com/gofiber/fiber/v3"
)

func AdminAuthMiddleware(c fiber.Ctx) error {

	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization format"})
	}

	claims, err := utils.ParseToken(parts[1])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	c.Locals("admin_id", claims["id"])
	c.Locals("email", claims["email"])

	return c.Next()
}
