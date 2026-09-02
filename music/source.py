from __future__ import annotations

import asyncio
import logging
from typing import Any

import yt_dlp
from jiosaavnpy import JioSaavn

from .queue import Track

log = logging.getLogger(__name__)

_YTDLP_BASE: dict[str, Any] = {
    "quiet": True,
    "no_warnings": True,
    "noplaylist": True,
    "skip_download": True,
    "extract_flat": False,
    "socket_timeout": 15,
    "retries": 2,
}


def _yt_extract(query: str, requester_id: int) -> Track:
    target = query if query.startswith(("http://", "https://")) else f"ytsearch1:{query}"
    options = dict(_YTDLP_BASE)
    options["format"] = "bestaudio/best"
    with yt_dlp.YoutubeDL(options) as ydl:
        info = ydl.extract_info(target, download=False)
        if not info:
            raise ValueError("No YouTube result found")
        if "entries" in info:
            entries = [entry for entry in info["entries"] if entry]
            if not entries:
                raise ValueError("No YouTube result found")
            info = entries[0]
        stream_url = info.get("url")
        if not stream_url:
            raise ValueError("YouTube stream URL unavailable")
        return Track((info.get("title") or "Unknown title").strip(), stream_url, info.get("webpage_url") or target, int(info["duration"]) if info.get("duration") else None, requester_id)


def _jio_extract(query: str, requester_id: int) -> Track:
    if query.startswith(("http://", "https://")):
        raise ValueError("JioSaavn search fallback requires a song query")
    results = JioSaavn().search_songs(query, limit=5)
    if not results:
        raise ValueError("No JioSaavn result found")
    for song in results:
        streams = song.get("stream_urls") or {}
        stream_url = streams.get("high_quality") or streams.get("medium_quality") or streams.get("low_quality")
        if not stream_url:
            continue
        try:
            duration = int(song.get("duration")) if song.get("duration") else None
        except (TypeError, ValueError):
            duration = None
        return Track((song.get("title") or query).strip(), stream_url, song.get("perma_url") or "https://www.jiosaavn.com/", duration, requester_id)
    raise ValueError("JioSaavn returned no playable stream")


def _sc_extract(query: str, requester_id: int) -> Track:
    target = query if query.startswith(("http://", "https://")) else f"scsearch1:{query}"
    options = dict(_YTDLP_BASE)
    options["format"] = "bestaudio/best"
    with yt_dlp.YoutubeDL(options) as ydl:
        info = ydl.extract_info(target, download=False)
        if not info:
            raise ValueError("No SoundCloud result found")
        if "entries" in info:
            entries = [entry for entry in info["entries"] if entry]
            if not entries:
                raise ValueError("No SoundCloud result found")
            info = entries[0]
        stream_url = info.get("url")
        if not stream_url:
            raise ValueError("SoundCloud stream URL unavailable")
        return Track((info.get("title") or "Unknown title").strip(), stream_url, info.get("webpage_url") or target, int(info["duration"]) if info.get("duration") else None, requester_id)


async def resolve(query: str, requester_id: int) -> Track:
    query = query.strip()
    if not query:
        raise ValueError("Empty search query")
    errors: list[str] = []
    for name, resolver in (("YouTube", _yt_extract), ("JioSaavn", _jio_extract), ("SoundCloud", _sc_extract)):
        try:
            return await asyncio.to_thread(resolver, query, requester_id)
        except Exception as exc:
            errors.append(f"{name}: {type(exc).__name__}")
            log.warning("%s source failed for %r: %s", name, query, exc)
    raise RuntimeError("All music sources failed: " + ", ".join(errors))
