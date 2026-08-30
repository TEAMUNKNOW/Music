# Architecture Deep Dive

## High-Level Flow

1. User opens bot → clicks button / deep link → Telegram opens Mini App
2. Mini App loads instantly (Svelte bundle is tiny)
3. Mini App sends `initData` to backend → HMAC verified
4. User searches / plays song
5. Backend returns signed short-lived stream URL + metadata
6. Frontend starts HLS / progressive playback in < 200ms
7. Optional: Join Sync Room → WebSocket keeps all clients in perfect sync

## Components

### 1. Telegram Mini App (Frontend)
- Runs inside Telegram WebView
- Uses `@twa-dev/sdk` for theme, haptic, user data
- Audio engine: Howler.js for gapless + Web Audio for visualizer
- State management: Svelte stores (or Zustand if React)

### 2. Go Backend
- HTTP API (Fiber preferred for speed)
- WebSocket hub for Sync Rooms
- Stream proxy / HLS generator
- Redis for:
  - Audio metadata cache
  - Hot audio chunks
  - Active rooms state

### 3. Streaming Pipeline
```
Search → Resolve source (YT / JioSaavn) → Cache check →
Generate signed URL → Frontend requests range/HLS →
Backend serves (or proxies) with proper headers
```

Key optimizations:
- Never download full file before streaming
- Prefer sources that already support range requests
- Aggressive Redis caching of popular tracks
- Pre-warm popular songs

### 4. Security Layer
- Every API call: validate Telegram `initData` HMAC-SHA256
- Stream URLs: HMAC signed + expiry (60-120 seconds)
- Rate limiting per user / IP
- No permanent public audio endpoints

### 5. Sync Room (Listen Together)
- WebSocket room per group / private session
- State: current track, position, isPlaying, queue
- Server is source of truth
- Clients send seek/pause/play → broadcast to room
- Drift correction every few seconds

## Performance Targets

- First byte: < 150ms (cached)
- Full UI interactive: < 800ms on mid-range Android
- 10,000 concurrent listeners: < 50MB RAM + low CPU
- Sync latency: < 80ms between clients
