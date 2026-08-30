package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CastRequest struct {
	ChatID   int64  `json:"chat_id"`
	TrackID  string `json:"track_id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
}

// CastToVC forwards a cast request to the userbot service
func CastToVC(c *fiber.Ctx) error {
	var req CastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.ChatID == 0 || req.TrackID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "chat_id and track_id required"})
	}

	// Resolve stream URL first
	info, err := searchResolve(req.TrackID)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "could not resolve track", "details": err.Error()})
	}

	castService := os.Getenv("CAST_SERVICE_URL")
	if castService == "" {
		castService = "http://localhost:4000"
	}

	payload := map[string]interface{}{
		"chat_id":    req.ChatID,
		"stream_url": info.URL,
		"title":      info.Title,
		"artist":     info.Artist,
		"track_id":   req.TrackID,
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(castService+"/cast", "application/json", bytes.NewReader(body))
	if err != nil {
		return c.Status(502).JSON(fiber.Map{
			"error":   "cast service unavailable",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "cast failed"})
	}

	return c.JSON(fiber.Map{
		"status":  "casting",
		"chat_id": req.ChatID,
		"title":   info.Title,
	})
}

// helper to avoid circular import issues in this file
func searchResolve(id string) (*struct {
	URL    string
	Title  string
	Artist string
}, error) {
	info, err := resolveStream(id)
	if err != nil {
		return nil, err
	}
	return &struct {
		URL    string
		Title  string
		Artist string
	}{URL: info.URL, Title: info.Title, Artist: info.Artist}, nil
}
