package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/TEAMUNKNOW/Music/internal/search"
	"github.com/TEAMUNKNOW/Music/internal/stream"
	"github.com/TEAMUNKNOW/Music/internal/sync"
)

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

// GetTrack returns metadata + signed stream URL
func GetTrack(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	signed := stream.GenerateSignedURL(id, baseURL)

	return c.JSON(fiber.Map{
		"id":        id,
		"streamUrl": signed,
		"expiresIn": 90,
	})
}

// Stream serves audio with signature validation + byte-range
func Stream(c *fiber.Ctx) error {
	id := c.Params("id")
	// In production, resolve real source URL from cache/DB using track ID
	// For now we return a clear message — real source resolution comes next
	sourceURL := c.Query("src") // temporary for testing
	if sourceURL == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "source URL resolution not yet connected — pass ?src= for testing",
		})
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
