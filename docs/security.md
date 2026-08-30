# Security Model

## 1. Telegram WebApp Authentication

Every request from Mini App must include the raw `initData` string.

Backend verification (Go):

```go
func ValidateInitData(initData, botToken string) (bool, map[string]string) {
    // 1. Parse query string
    // 2. Extract hash
    // 3. Sort remaining key=value
    // 4. HMAC-SHA256 with secret key = HMAC-SHA256("WebAppData", botToken)
    // 5. Compare
}
```

- Reject if hash mismatch
- Reject if `auth_date` older than 24h (or stricter)
- Extract `user.id` safely after verification

## 2. Signed Streaming URLs

Never expose permanent audio links.

Example signed URL:
```
/stream/{trackID}?exp=1725000000&sig=abc123...
```

- `exp` = unix timestamp (now + 60~120s)
- `sig` = HMAC-SHA256 of `trackID + exp` using a strong secret

On request:
1. Check expiry
2. Recompute signature
3. Serve only if valid

This prevents bandwidth theft and hotlinking.

## 3. Rate Limiting

- Per Telegram user ID
- Per IP (secondary)
- Search endpoints more strict than stream

## 4. Additional Hardening

- CORS locked to Telegram origins only
- No debug endpoints in production
- Secrets only via environment variables
- Redis AUTH enabled
- Minimal attack surface (no unnecessary packages)
