package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
)

func RegisterWharehouseRoute(app *fiber.App, q *db.Queries, conn *pgxpool.Pool) {
	wharehouses := app.Group("/wharehouses")

	wharehouses.Get("/", func(ctx fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")

		if authHeader != "KKLLXXZ" {
			return ctx.Status(500).JSON(fiber.Map{"error": "Not Auth"})
		}
		res, err := q.ListWarehouses(ctx.Context())
		if err != nil {
			return ctx.Status(500).JSON(fiber.Map{"error": err})
		}
		return ctx.JSON(res)
	})
}
