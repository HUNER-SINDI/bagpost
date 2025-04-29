package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
)

func SetupRouters(app *fiber.App, q *db.Queries, conn *pgxpool.Pool) {
	RegisterWharehouseRoute(app, q, conn)
	RegisterAdminRoute(app, q, conn)
}
