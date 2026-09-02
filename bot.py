from __future__ import annotations

import asyncio
import logging
import time

from pyrogram import Client, filters
from pyrogram.enums import ChatMemberStatus
from pyrogram.types import Message
from pytgcalls import PyTgCalls, filters as call_filters
from pytgcalls.types import StreamEnded

from config import Config
from music.player import MusicPlayer
from music.queue import QueueManager
from music.source import resolve

logging.basicConfig(level=logging.INFO, format="%(asctime)s | %(levelname)s | %(name)s | %(message)s")
log = logging.getLogger("music")

config = Config()

app = Client(
    "music-bot",
    api_id=config.API_ID,
    api_hash=config.API_HASH,
    bot_token=config.BOT_TOKEN,
)
calls = PyTgCalls(app)
queues = QueueManager()
player = MusicPlayer(calls, queues)


async def admin_only(message: Message) -> bool:
    if not message.from_user or message.chat.type.value == "private":
        return True
    member = await app.get_chat_member(message.chat.id, message.from_user.id)
    return member.status in {ChatMemberStatus.OWNER, ChatMemberStatus.ADMINISTRATOR}


@app.on_message(filters.command("start"))
async def start_handler(_: Client, message: Message) -> None:
    await message.reply_text(
        "🎵 **Music Bot**\n\n"
        "`/play <song name or URL>` — play audio\n"
        "`/queue` — show queue\n"
        "`/pause` / `/resume` — control playback\n"
        "`/skip` — skip current track\n"
        "`/stop` — stop and clear queue\n"
        "`/volume 0-200` — set volume\n"
        "`/ping` — health check"
    )


@app.on_message(filters.command("ping"))
async def ping_handler(_: Client, message: Message) -> None:
    started = time.perf_counter()
    reply = await message.reply_text("🏓 Checking…")
    elapsed = (time.perf_counter() - started) * 1000
    await reply.edit_text(f"🏓 **Pong:** `{elapsed:.0f} ms`")


@app.on_message(filters.command("play"))
async def play_handler(_: Client, message: Message) -> None:
    if len(message.command) < 2:
        await message.reply_text("Usage: `/play <song name or URL>`")
        return
    query = message.text.split(None, 1)[1].strip()
    requester = message.from_user.id if message.from_user else 0
    status = await message.reply_text("🔎 Searching and preparing audio…")
    try:
        track = await resolve(query, requester)
        started, position = await player.queue_or_play(message.chat.id, track)
        if started:
            await status.edit_text(f"▶️ **Playing:** [{track.title}]({track.webpage_url})", disable_web_page_preview=True)
        else:
            await status.edit_text(f"➕ **Queued #{position}:** [{track.title}]({track.webpage_url})", disable_web_page_preview=True)
    except Exception as exc:
        log.exception("Play failed")
        await status.edit_text(f"❌ Could not play that source.\n`{type(exc).__name__}`")


@app.on_message(filters.command("queue"))
async def queue_handler(_: Client, message: Message) -> None:
    current, pending = await queues.snapshot(message.chat.id)
    if current is None and not pending:
        await message.reply_text("📭 Queue is empty.")
        return
    lines = [f"▶️ **Now:** {current.title}" if current else "▶️ **Now:** —"]
    if pending:
        lines.append("\n📋 **Up next:**")
        lines.extend(f"`{i}.` {track.title}" for i, track in enumerate(pending, 1))
    await message.reply_text("\n".join(lines))


@app.on_message(filters.command(["pause", "resume", "skip", "stop", "volume"]))
async def control_handler(_: Client, message: Message) -> None:
    if not await admin_only(message):
        await message.reply_text("❌ Admin permission required.")
        return
    command = message.command[0].lower()
    try:
        if command == "pause":
            await player.pause(message.chat.id)
            await message.reply_text("⏸️ Paused.")
        elif command == "resume":
            await player.resume(message.chat.id)
            await message.reply_text("▶️ Resumed.")
        elif command == "skip":
            track = await player.skip(message.chat.id)
            await message.reply_text(f"⏭️ **Playing:** {track.title}" if track else "⏹️ Queue finished.")
        elif command == "stop":
            await player.stop(message.chat.id)
            await message.reply_text("⏹️ Stopped and cleared the queue.")
        elif command == "volume":
            if len(message.command) != 2 or not message.command[1].isdigit():
                await message.reply_text("Usage: `/volume 0-200`")
                return
            await player.volume(message.chat.id, int(message.command[1]))
            await message.reply_text(f"🔊 Volume set to `{message.command[1]}`.")
    except Exception as exc:
        log.exception("Control command failed")
        await message.reply_text(f"❌ Command failed: `{type(exc).__name__}`")


@calls.on_update(call_filters.stream_end())
async def stream_end_handler(_: PyTgCalls, update: StreamEnded) -> None:
    chat_id = update.chat_id
    try:
        if await queues.was_recent_manual_action(chat_id, time.monotonic()):
            return
        state = await queues.state(chat_id)
        state.current = None
        track = await player.start_next(chat_id)
        if track:
            log.info("Auto-advanced in %s: %s", chat_id, track.title)
        else:
            await player.stop(chat_id)
    except Exception:
        log.exception("Automatic queue advance failed in %s", chat_id)


async def main() -> None:
    calls.start()
    me = await app.get_me()
    log.info("Music bot started as @%s", me.username or me.id)
    try:
        await asyncio.Event().wait()
    finally:
        await app.stop()


if __name__ == "__main__":
    asyncio.run(main())
