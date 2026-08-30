# Music

**Spotify inside Telegram** — Hybrid Telegram Mini App (TMA) + High-Performance Go Streaming Backend.

Zero chat spam • Sub-200ms buffering • Real-time Listen Together • Fully secured

---

## Why this is different from Yukki / AnonX / DaisyX forks

| Feature                    | Common Bots          | This Project                  |
|---------------------------|----------------------|-------------------------------|
| UI                        | Chat + Inline buttons| Native Telegram Mini App      |
| Chat Spam                 | Heavy                | Zero                          |
| Buffering                 | Full download + FFmpeg | Instant HLS / Byte-Range     |
| Listen Together           | None / basic         | Real-time WebSocket Sync Rooms|
| Performance               | High RAM/CPU         | ~40MB for 10k listeners       |
| Security                  | Weak / none          | HMAC + Signed Stream URLs     |

---

## Architecture Overview

```
[ Telegram Client ]
        │
        ▼  (opens Mini App)
[ Svelte / React TMA ]  ←── WebSocket + HMAC verified API
        │
        ▼
[ Go Backend (Fiber) ]
        │
   ┌────┴────┐
   ▼         ▼
[ Redis ]  [ HLS / Byte-Range Streamer ]
```

### Modes
1. **Personal Mode** — High quality 320kbps streaming inside Mini App (no VC needed)
2. **Group Voice Chat Mode** — "Cast to VC" button → userbot streams to Telegram Voice Chat

---

## Tech Stack

### Frontend (Telegram Mini App)
- **Framework**: Svelte + Vite (or React)
- **Styling**: Tailwind CSS
- **Audio**: Howler.js / Web Audio API
- **Telegram SDK**: `@twa-dev/sdk`
- Features: Haptic feedback, theme sync, 60/120 FPS UI, live waveform, LRC lyrics, swipe gestures

### Backend
- **Language**: Go (Fiber / Gin)
- **Streaming**: HTTP Byte-Range + HLS chunking
- **Cache**: Redis (in-memory audio + metadata)
- **Auth**: Telegram WebApp `initData` HMAC-SHA256 verification
- **Stream Security**: Time-limited signed URLs (1-min expiry)
- **Search**: YouTube (Piped / Invidious) + JioSaavn metadata

### Optional
- Userbot for Group Voice Chat casting (Pyrogram / Telethon or Go equivalent)

---

## Project Structure

```
Music/
├── frontend/                 # Telegram Mini App (Svelte + Vite)
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── lib/
│   │   └── stores/
│   ├── public/
│   └── package.json
├── backend/                  # Go streaming + API server
│   ├── cmd/
│   ├── internal/
│   │   ├── auth/
│   │   ├── stream/
│   │   ├── search/
│   │   ├── sync/
│   │   └── cache/
│   ├── go.mod
│   └── main.go
├── docs/
│   ├── architecture.md
│   ├── api.md
│   └── security.md
├── docker-compose.yml
└── README.md
```

---

## Key Features Roadmap

- [x] Architecture design
- [ ] Mini App UI (Home, Player, Search, Sync Room)
- [ ] Go backend skeleton + HMAC auth
- [ ] Instant HLS / Byte-Range streaming
- [ ] Redis caching layer
- [ ] Real-time WebSocket sync rooms
- [ ] Signed streaming URLs
- [ ] Cast to Telegram Voice Chat
- [ ] Live waveform + synchronized lyrics

---

## Getting Started (coming soon)

```bash
# Frontend
cd frontend && npm install && npm run dev

# Backend
cd backend && go run .
```

---

## Security Notes

- All API requests must include valid Telegram `initData`
- Backend verifies HMAC-SHA256 using Bot Token
- Audio stream links are short-lived signed tokens
- No permanent public audio URLs

---

**Status**: Scaffolding in progress  
Built for speed, security, and native Telegram experience.
