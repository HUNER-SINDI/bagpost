package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
	"toxkurd.com/bagpost/middleware"
	utils "toxkurd.com/bagpost/util/warehouse"
)

func RegisterWharehouseRoute(app *fiber.App, q *db.Queries, conn *pgxpool.Pool) {
	wharehouses := app.Group("/warehouses")

	wharehouses.Get("/", func(ctx fiber.Ctx) error {
		res, err := q.ListWarehouses(ctx.Context())
		if err != nil {
			return ctx.Status(500).JSON(fiber.Map{"error": err})
		}
		return ctx.JSON(res)
	}, middleware.WarehouseAuthMiddleware)

	wharehouses.Post("/login", func(c fiber.Ctx) error {
		type LoginInput struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var input LoginInput
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Json input"})
		}

		// Fetch warehouse user by email
		warehouseUser, err := q.GetWarehouseByEmail(c.Context(), input.Email)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}

		// Check if account is active
		if !warehouseUser.IsActive.Bool {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is inactive"})
		}

		// Compare password
		if warehouseUser.Password != input.Password {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}

		// Generate JWT and Refresh Token
		accessToken, err := utils.GenerateToken(int(warehouseUser.ID), warehouseUser.Email, warehouseUser.IsActive.Bool)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		refreshToken, err := utils.GenerateRefreshToken(int(warehouseUser.ID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate refresh token"})
		}

		// Respond with tokens
		return c.JSON(fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	wharehouses.Post("/refresh", func(c fiber.Ctx) error {
		type RefreshTokenInput struct {
			RefreshToken string `json:"refresh_token"`
		}

		var input RefreshTokenInput
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		// Parse and validate the refresh token
		warehouseID, err := utils.ParseRefreshToken(input.RefreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
		}

		// Fetch the warehouse user by ID
		warehouseUser, err := q.GetWarehouseByID(c.Context(), int32(warehouseID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch warehouse user"})
		}

		// Check if the account is active
		if !warehouseUser.IsActive.Bool {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is inactive"})
		}

		// Generate a new access token
		accessToken, err := utils.GenerateToken(int(warehouseUser.ID), warehouseUser.Email, warehouseUser.IsActive.Bool)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate new access token"})
		}

		// Generate a new refresh token (with a longer expiration time)
		refreshToken, err := utils.GenerateRefreshToken(int(warehouseUser.ID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate new refresh token"})
		}

		// Respond with both new access token and refresh token
		return c.JSON(fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

}
