package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/TEAMUNKNOW/Music/internal/search"
)

type CastRequest struct {
	ChatID  int64  `json:"chat_id"`
	TrackID string `json:"track_id"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
}

func castServiceURL() string {
	u := os.Getenv("CAST_SERVICE_URL")
	if u == "" {
		return "http://localhost:4000"
	}
	return u
}

// CastStatus reports whether Cast-to-VC is available (session configured)
func CastStatus(c *fiber.Ctx) error {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(castServiceURL() + "/health")
	if err != nil {
		return c.JSON(fiber.Map{
			"cast_enabled": false,
			"reason":       "cast service unreachable or not running",
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var health struct {
		CastEnabled bool `json:"cast_enabled"`
		Ready       bool `json:"ready"`
	}
	_ = json.Unmarshal(body, &health)

	enabled := health.CastEnabled || health.Ready
	return c.JSON(fiber.Map{
		"cast_enabled": enabled,
		"reason":       ternary(enabled, "userbot session active", "no session configured"),
	})
}

// CastToVC forwards a cast request only if the userbot session is active
func CastToVC(c *fiber.Ctx) error {
	var req CastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.ChatID == 0 || req.TrackID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "chat_id and track_id required"})
	}

	// Quick check if cast is even available
	statusClient := &http.Client{Timeout: 2 * time.Second}
	hResp, err := statusClient.Get(castServiceURL() + "/health")
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"error":   "Cast-to-VC is disabled",
			"reason":  "cast service not running — this feature is optional",
			"hint":    "Provide SESSION_STRING + API_ID/API_HASH to enable",
		})
	}
	defer hResp.Body.Close()
	hBody, _ := io.ReadAll(hResp.Body)
	var health struct {
		CastEnabled bool `json:"cast_enabled"`
		Ready       bool `json:"ready"`
	}
	_ = json.Unmarshal(hBody, &health)
	if !health.CastEnabled && !health.Ready {
		return c.Status(503).JSON(fiber.Map{
			"error":  "Cast-to-VC is disabled",
			"reason": "no userbot session configured",
			"hint":   "Set SESSION_STRING (and API_ID/API_HASH) in cast-service to enable",
		})
	}

	info, err := search.ResolveStream(req.TrackID, "")
	if err != nil {
		return c.Status(502).JSON(fiber.Map{
			"error":   "could not resolve track",
			"details": err.Error(),
		})
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
	resp, err := client.Post(castServiceURL()+"/cast", "application/json", bytes.NewReader(body))
	if err != nil {
		return c.Status(502).JSON(fiber.Map{
			"error":   "cast service error",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).JSON(fiber.Map{
			"error":  "cast failed",
			"detail": string(respBody),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "casting",
		"chat_id": req.ChatID,
		"title":   info.Title,
	})
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
