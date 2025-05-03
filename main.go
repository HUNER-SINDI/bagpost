package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
	"toxkurd.com/bagpost/router"
)

func main() {
	dsn := "postgres://postgres:Huner@212@localhost:5432/bagpost?sslmode=disable"
	// dsn := os.Getenv("DATABASE_URL")
	// if dsn == "" {
	// 	log.Fatal("DATABASE_URL is not set")
	// }
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("cannot connect to server: %v", err)
	}
	defer pool.Close()

	app := fiber.New()
	// 🔥 Enable CORS for all origins
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // ✅ Must be a slice of strings
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))
	q := db.New(pool)

	router.SetupRouters(app, q, pool)

	log.Println("Server running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
