package handlers

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"

	"github.com/TEAMUNKNOW/Music/internal/search"
	"github.com/TEAMUNKNOW/Music/internal/stream"
	"github.com/TEAMUNKNOW/Music/internal/sync"
)

var redisClient *redis.Client

func InitRedis(addr string) {
	redisClient = redis.NewClient(&redis.Options{Addr: addr})
}

// Search handles music search requests
func Search(c *fiber.Ctx) error {
	q := c.Query("q")
	if q == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query required"})
	}

	tracks, err := search.Search(q)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "search failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"query":   q,
		"results": tracks,
	})
}

// GetTrack resolves real audio URL, caches source, returns signed proxy URL + metadata
func GetTrack(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	source := c.Query("source") // youtube | jiosaavn

	info, err := search.ResolveStream(id, source)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{
			"error":   "failed to resolve stream",
			"details": err.Error(),
		})
	}

	// Cache the real source URL for the stream handler (5 min)
	if redisClient != nil {
		ctx := context.Background()
		redisClient.Set(ctx, "src:"+id, info.URL, 5*time.Minute)
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	signed := stream.GenerateSignedURL(id, baseURL)

	return c.JSON(fiber.Map{
		"id":        id,
		"title":     info.Title,
		"artist":    info.Artist,
		"duration":  info.Duration,
		"thumbnail": info.Thumbnail,
		"quality":   info.Quality,
		"streamUrl": signed,
		"expiresIn": 90,
	})
}

// Stream serves audio with signature validation + byte-range proxy
func Stream(c *fiber.Ctx) error {
	id := c.Params("id")
	exp := c.Query("exp")
	sig := c.Query("sig")

	if !stream.ValidateSignedRequest(id, exp, sig) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid or expired signature"})
	}

	var sourceURL string
	if redisClient != nil {
		sourceURL, _ = redisClient.Get(context.Background(), "src:"+id).Result()
	}

	// Fallback: allow ?src= for testing
	if sourceURL == "" {
		sourceURL = c.Query("src")
	}
	if sourceURL == "" {
		// Last resort: try resolve again
		info, err := search.ResolveStream(id, "")
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "stream source not found"})
		}
		sourceURL = info.URL
	}

	return stream.ServeStream(c, id, sourceURL)
}

// WebSocketUpgrade upgrades to WebSocket for Sync Rooms
func WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return websocket.New(handleWS)(c)
	}
	return fiber.ErrUpgradeRequired
}

func handleWS(conn *websocket.Conn) {
	defer conn.Close()

	roomID := conn.Query("room")
	if roomID == "" {
		roomID = "default"
	}
	userID := conn.Query("user")
	if userID == "" {
		userID = "anon"
	}

	room := sync.GetOrCreateRoom(roomID)
	room.Join(conn, userID)
	defer room.Leave(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		room.HandleMessage(msg)
	}
}
