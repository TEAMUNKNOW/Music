# Cast to Group Voice Chat

## Goal
User is listening inside Mini App → taps **Cast to Group Voice Chat** → song starts playing in the Telegram group’s Voice Chat for everyone.

## Architecture

```
Mini App
   │  POST /api/cast
   ▼
Go Backend
   │  (validates user + track)
   ▼
Userbot / Stream Bridge
   │  joins group VC as a participant
   │  streams audio (raw PCM or Opus)
   ▼
Telegram Group Voice Chat
```

## Implementation Options

### Option A – Python Userbot (fastest to ship)
- Pyrogram / Telethon + pytgcalls
- Backend calls an internal HTTP endpoint on the userbot service
- Userbot joins the chat’s VC and plays the audio URL

### Option B – Pure Go (harder)
- Use gotd + custom voice streaming (complex Opus handling)

### Recommended for v1
Use a small separate Python service:

```
cast-service/
  ├── main.py          # FastAPI or aiohttp
  ├── player.py        # pytgcalls wrapper
  └── requirements.txt
```

Backend endpoint:

```
POST /api/cast
{
  "chat_id": -100xxxxxxxxxx,
  "track_id": "...",
  "stream_url": "signed-url"
}
```

## Security
- Only allow cast if the requesting user is admin / has permission in that group
- Rate-limit cast requests
- Signed stream URL still required

## Status
Scaffold ready. Userbot service + endpoint wiring is the next concrete task.
