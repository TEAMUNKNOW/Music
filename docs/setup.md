# Setup Guide

## 1. Backend + Redis

```bash
cp .env.example .env
# Fill BOT_TOKEN and STREAM_SECRET

docker compose up -d redis backend
# or
cd backend && go mod tidy && go run .
```

## 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

Set `VITE_API_BASE=http://localhost:3000` if needed.

## 3. Cast Service (optional)

```bash
cd cast-service
pip install -r requirements.txt

# Get API_ID + API_HASH from https://my.telegram.org
# Generate SESSION_STRING once (use pyrogram string session generator)

export API_ID=...
export API_HASH=...
export SESSION_STRING=...

python main.py
```

Userbot account must be admin (or have permission) in the target group to join Voice Chat.

## How playback works now

1. User searches → Piped / JioSaavn results
2. User taps song → `GET /api/track/:id?source=...`
3. Backend resolves real audio URL (Piped streams / JioSaavn 320kbps)
4. Caches source in Redis + returns **signed** `/stream/:id?exp=&sig=`
5. Frontend plays the signed URL via Howler (byte-range proxy)
6. Stream handler validates signature and proxies the real audio

## Cast flow

1. User opens Mini App inside a group
2. Plays a song → taps **Cast to Group Voice Chat**
3. Frontend sends `chat_id` + `track_id` to `/api/cast`
4. Backend resolves stream → calls cast-service
5. Userbot joins the group VC and plays the audio
