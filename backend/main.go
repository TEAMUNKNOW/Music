package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/TEAMUNKNOW/Music/internal/auth"
	"github.com/TEAMUNKNOW/Music/internal/handlers"
)

func main() {
	_ = godotenv.Load()

	app := fiber.New(fiber.Config{
		AppName: "Music Streaming Backend",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://web.telegram.org,https://*.telegram.org",
		AllowHeaders: "Origin, Content-Type, Accept, X-Telegram-Init-Data",
	}))

	// Public health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Protected API group
	api := app.Group("/api", auth.TelegramAuthMiddleware)

	api.Get("/search", handlers.Search)
	api.Get("/track/:id", handlers.GetTrack)
	api.Get("/stream/:id", handlers.Stream) // signed URL preferred, this is fallback

	// WebSocket for Sync Rooms
	app.Get("/ws", handlers.WebSocketUpgrade)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
