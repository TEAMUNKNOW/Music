package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TelegramAuthMiddleware(c *fiber.Ctx) error {
	initData := c.Get("X-Telegram-Init-Data")
	if initData == "" { initData = c.Query("initData") }
	if initData == "" { return c.Status(401).JSON(fiber.Map{"error":"missing initData"}) }
	botToken := strings.TrimSpace(os.Getenv("BOT_TOKEN"))
	if botToken == "" { return c.Status(500).JSON(fiber.Map{"error":"server misconfigured"}) }
	valid, userID := validateInitData(initData, botToken)
	if !valid { return c.Status(401).JSON(fiber.Map{"error":"invalid initData"}) }
	c.Locals("userID", userID)
	return c.Next()
}

func validateInitData(initData, botToken string) (bool, int64) {
	values, err := url.ParseQuery(initData)
	if err != nil { return false, 0 }
	hash := values.Get("hash")
	if hash == "" { return false, 0 }
	values.Del("hash")

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil { return false, 0 }
	now := time.Now().Unix()
	if authDate > now+60 || now-authDate > 86400 { return false, 0 }

	pairs := make([]string, 0, len(values))
	for k, v := range values { if len(v)>0 { pairs=append(pairs,k+"="+v[0]) } }
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs,"\n")

	// Telegram Web Apps: secret_key = HMAC-SHA256(key=bot_token, data="WebAppData")
	secretMAC := hmac.New(sha256.New, []byte(botToken))
	_, _ = secretMAC.Write([]byte("WebAppData"))
	secretKey := secretMAC.Sum(nil)

	checkMAC := hmac.New(sha256.New, secretKey)
	_, _ = checkMAC.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(checkMAC.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(hash))) { return false, 0 }

	var user struct { ID int64 `json:"id"` }
	if raw := values.Get("user"); raw != "" { _ = json.Unmarshal([]byte(raw), &user) }
	return true, user.ID
}
