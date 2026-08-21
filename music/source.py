from __future__ import annotations

import asyncio
from typing import Any

import yt_dlp

from .queue import Track


_YTDLP_BASE: dict[str, Any] = {
    "quiet": True,
    "no_warnings": True,
    "noplaylist": True,
    "skip_download": True,
    "extract_flat": False,
    "socket_timeout": 15,
    "retries": 2,
}


def _extract(query: str, requester_id: int) -> Track:
    target = query if query.startswith(("http://", "https://")) else f"ytsearch1:{query}"
    options = dict(_YTDLP_BASE)
    options["format"] = "bestaudio/best"
    with yt_dlp.YoutubeDL(options) as ydl:
        info = ydl.extract_info(target, download=False)
        if not info:
            raise ValueError("No result found")
        if "entries" in info:
            entries = [entry for entry in info["entries"] if entry]
            if not entries:
                raise ValueError("No result found")
            info = entries[0]
        stream_url = info.get("url")
        if not stream_url:
            raise ValueError("Audio stream URL was not available")
        return Track(
            title=(info.get("title") or "Unknown title").strip(),
            stream_url=stream_url,
            webpage_url=info.get("webpage_url") or target,
            duration=int(info["duration"]) if info.get("duration") else None,
            requester_id=requester_id,
        )


async def resolve(query: str, requester_id: int) -> Track:
    query = query.strip()
    if not query:
        raise ValueError("Empty search query")
    return await asyncio.to_thread(_extract, query, requester_id)
