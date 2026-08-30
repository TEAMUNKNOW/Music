package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Search handles music search requests
func Search(c *fiber.Ctx) error {
	q := c.Query("q")
	if q == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query required"})
	}

	// TODO: implement Piped / Invidious / JioSaavn search + Redis cache
	return c.JSON(fiber.Map{
		"query":   q,
		"results": []interface{}{},
		"message": "search stub – implement source aggregator",
	})
}

// GetTrack returns metadata for a track
func GetTrack(c *fiber.Ctx) error {
	id := c.Params("id")
	// TODO: fetch from cache or source
	return c.JSON(fiber.Map{
		"id":       id,
		"title":    "Placeholder Track",
		"artist":   "Unknown",
		"duration": 210,
		"message":  "track stub",
	})
}

// Stream serves audio (prefer signed URLs in production)
func Stream(c *fiber.Ctx) error {
	id := c.Params("id")
	// TODO: validate signed URL params, serve byte-range or HLS
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"id":      id,
		"message": "streaming engine coming soon – use signed URLs",
	})
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
	// TODO: room join / leave / state broadcast logic
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Echo for now
		_ = conn.WriteMessage(mt, msg)
	}
}
