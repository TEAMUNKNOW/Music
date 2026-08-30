package stream

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

const defaultExpiry = 90 // seconds

// GenerateSignedURL creates a time-limited signed stream URL
func GenerateSignedURL(trackID string, baseURL string) string {
	exp := time.Now().Unix() + defaultExpiry
	sig := sign(trackID, exp)
	return fmt.Sprintf("%s/stream/%s?exp=%d&sig=%s", baseURL, trackID, exp, sig)
}

// ValidateSignedRequest checks exp + signature
func ValidateSignedRequest(trackID string, expStr, sig string) bool {
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > exp {
		return false
	}
	expected := sign(trackID, exp)
	return hmac.Equal([]byte(expected), []byte(sig))
}

func sign(trackID string, exp int64) string {
	secret := os.Getenv("STREAM_SECRET")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%d", trackID, exp)))
	return hex.EncodeToString(mac.Sum(nil))
}
