package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TelegramAuthMiddleware validates Telegram WebApp initData
func TelegramAuthMiddleware(c *fiber.Ctx) error {
	initData := c.Get("X-Telegram-Init-Data")
	if initData == "" {
		initData = c.Query("initData")
	}
	if initData == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing initData",
		})
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "server misconfigured",
		})
	}

	valid, userID := validateInitData(initData, botToken)
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid initData",
		})
	}

	c.Locals("userID", userID)
	return c.Next()
}

func validateInitData(initData, botToken string) (bool, int64) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return false, 0
	}

	hash := values.Get("hash")
	if hash == "" {
		return false, 0
	}
	values.Del("hash")

	// Check auth_date freshness (24h)
	authDateStr := values.Get("auth_date")
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil || time.Now().Unix()-authDate > 86400 {
		return false, 0
	}

	// Build data-check-string
	var pairs []string
	for k, v := range values {
		if len(v) > 0 {
			pairs = append(pairs, k+"="+v[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// secret_key = HMAC_SHA256("WebAppData", botToken)
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return false, 0
	}

	// Extract user id if present
	var userID int64
	if userJSON := values.Get("user"); userJSON != "" {
		// Simple extraction (production: proper JSON parse)
		if strings.Contains(userJSON, "\"id\":") {
			// fallback simple parse
		}
	}

	return true, userID
}
