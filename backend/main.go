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
	"github.com/TEAMUNKNOW/Music/internal/search"
	"github.com/TEAMUNKNOW/Music/internal/stream"
)

func main() {
	_ = godotenv.Load()

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if len(redisAddr) > 8 && redisAddr[:8] == "redis://" {
		redisAddr = redisAddr[8:]
	}

	stream.InitRedis(redisAddr)
	search.Init(redisAddr)
	handlers.InitRedis(redisAddr)

	app := fiber.New(fiber.Config{
		AppName: "Music Streaming Backend",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, X-Telegram-Init-Data",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Public stream endpoint (protected by signature)
	app.Get("/stream/:id", handlers.Stream)

	// Protected API
	api := app.Group("/api", auth.TelegramAuthMiddleware)
	api.Get("/search", handlers.Search)
	api.Get("/track/:id", handlers.GetTrack)
	api.Post("/cast", handlers.CastToVC)

	// WebSocket Sync Rooms
	app.Get("/ws", handlers.WebSocketUpgrade)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Music backend listening on :%s", port)
	log.Fatal(app.Listen(":" + port))
}
