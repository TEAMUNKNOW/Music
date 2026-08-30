package stream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis(addr string) {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

// ServeStream handles authenticated + signed range requests
func ServeStream(c *fiber.Ctx, trackID, sourceURL string) error {
	exp := c.Query("exp")
	sig := c.Query("sig")

	if !ValidateSignedRequest(trackID, exp, sig) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid or expired signature"})
	}

	// Try Redis cache first for small popular tracks
	if rdb != nil {
		cached, err := rdb.Get(context.Background(), "audio:"+trackID).Bytes()
		if err == nil && len(cached) > 0 {
			return serveCached(c, cached)
		}
	}

	// Proxy with range support
	req, err := http.NewRequestWithContext(c.Context(), "GET", sourceURL, nil)
	if err != nil {
		return c.Status(500).SendString("failed to create request")
	}

	// Forward Range header if present
	if rangeHeader := c.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(502).SendString("upstream error")
	}
	defer resp.Body.Close()

	// Copy important headers
	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	if resp.Header.Get("Content-Length") != "" {
		c.Set("Content-Length", resp.Header.Get("Content-Length"))
	}
	if resp.Header.Get("Accept-Ranges") != "" {
		c.Set("Accept-Ranges", resp.Header.Get("Accept-Ranges"))
	}
	if resp.Header.Get("Content-Range") != "" {
		c.Set("Content-Range", resp.Header.Get("Content-Range"))
	}
	c.Status(resp.StatusCode)

	_, err = io.Copy(c.Response().BodyWriter(), resp.Body)
	return err
}

func serveCached(c *fiber.Ctx, data []byte) error {
	c.Set("Content-Type", "audio/mpeg")
	c.Set("Accept-Ranges", "bytes")
	c.Set("Content-Length", strconv.Itoa(len(data)))

	rangeHeader := c.Get("Range")
	if rangeHeader == "" {
		return c.Send(data)
	}

	// Simple bytes=0- implementation
	parts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(parts) != 2 {
		return c.Send(data)
	}

	start, _ := strconv.Atoi(parts[0])
	end := len(data) - 1
	if parts[1] != "" {
		end, _ = strconv.Atoi(parts[1])
	}
	if start < 0 || end >= len(data) || start > end {
		return c.Status(416).SendString("invalid range")
	}

	c.Status(206)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
	c.Set("Content-Length", strconv.Itoa(end-start+1))
	return c.Send(data[start : end+1])
}
