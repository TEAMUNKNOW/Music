# 🎵 Music Bot

A clean, async-first Telegram music bot focused on reliable voice-chat playback.

## Current core

- Per-chat isolated queues
- YouTube/search and direct URL resolution through `yt-dlp`
- Non-blocking source extraction with `asyncio.to_thread`
- Play / pause / resume / skip / stop / volume
- Automatic queue advance on `stream_end`
- Admin-only playback controls in groups
- Automatic cleanup when a queue finishes
- Defensive error handling and structured logging
- Docker + FFmpeg runtime
- Async queue tests

## Commands

- `/play <song or URL>`
- `/queue`
- `/pause`
- `/resume`
- `/skip`
- `/stop`
- `/volume 0-200`
- `/ping`

## Configuration

Copy `.env.example` to `.env` and set:

```text
API_ID=...
API_HASH=...
BOT_TOKEN=...
```

Never commit real credentials or Telegram session files.

## Run

```bash
pip install -r requirements.txt
python bot.py
```

## Docker

```bash
docker build -t music-bot .
docker run --env-file .env music-bot
```

## Reliability notes

The bot keeps queue state isolated by chat and serializes playback-changing operations per chat. Audio-source extraction is moved off the event loop so a slow resolver does not block unrelated Telegram updates. PyTgCalls provides the Telegram voice-chat media layer, including `MediaStream`, stream-end updates, and playback controls. citeturn0search5turn2search0

This project aims for production-grade stability, but no networked media bot can honestly guarantee zero failures: upstream media URLs, Telegram voice chats, FFmpeg, hosting, and external services can all fail. The code therefore treats failure recovery as a first-class concern.
