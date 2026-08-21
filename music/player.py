from __future__ import annotations

import asyncio
import logging
import time

from pytgcalls import PyTgCalls
from pytgcalls.types import MediaStream

from .queue import QueueManager, Track

log = logging.getLogger(__name__)


class MusicPlayer:
    def __init__(self, calls: PyTgCalls, queues: QueueManager) -> None:
        self.calls = calls
        self.queues = queues
        self._locks: dict[int, asyncio.Lock] = {}

    def _lock(self, chat_id: int) -> asyncio.Lock:
        return self._locks.setdefault(chat_id, asyncio.Lock())

    async def play_track(self, chat_id: int, track: Track) -> None:
        await self.calls.play(
            chat_id,
            MediaStream(track.stream_url, video_flags=MediaStream.Flags.IGNORE),
        )

    async def start_next(self, chat_id: int) -> Track | None:
        lock = self._lock(chat_id)
        async with lock:
            state = await self.queues.state(chat_id)
            if state.current is not None:
                return state.current
            track = await self.queues.pop_next(chat_id)
            if track is None:
                return None
            state.current = track
            try:
                await self.play_track(chat_id, track)
                return track
            except Exception:
                log.exception("Failed to start track in %s", chat_id)
                state.current = None
                return await self.start_next(chat_id)

    async def queue_or_play(self, chat_id: int, track: Track) -> tuple[bool, int]:
        lock = self._lock(chat_id)
        async with lock:
            state = await self.queues.state(chat_id)
            if state.current is None:
                state.current = track
                try:
                    await self.play_track(chat_id, track)
                    return True, 0
                except Exception:
                    state.current = None
                    raise
            position = await self.queues.enqueue(chat_id, track)
            return False, position

    async def skip(self, chat_id: int) -> Track | None:
        lock = self._lock(chat_id)
        async with lock:
            state = await self.queues.state(chat_id)
            await self.queues.mark_manual_action(chat_id, time.monotonic())
            state.current = None
            try:
                track = await self.queues.pop_next(chat_id)
                if track is None:
                    await self.calls.leave_call(chat_id)
                    return None
                state.current = track
                await self.play_track(chat_id, track)
                return track
            except Exception:
                state.current = None
                log.exception("Skip failed in %s", chat_id)
                return await self.start_next(chat_id)

    async def stop(self, chat_id: int) -> None:
        lock = self._lock(chat_id)
        async with lock:
            await self.queues.mark_manual_action(chat_id, time.monotonic())
            await self.queues.clear(chat_id)
            try:
                await self.calls.leave_call(chat_id)
            except Exception:
                log.debug("Voice call already closed for %s", chat_id, exc_info=True)

    async def pause(self, chat_id: int) -> None:
        await self.calls.pause(chat_id)

    async def resume(self, chat_id: int) -> None:
        await self.calls.resume(chat_id)

    async def volume(self, chat_id: int, value: int) -> None:
        if not 0 <= value <= 200:
            raise ValueError("Volume must be between 0 and 200")
        await self.calls.change_volume_call(chat_id, value)
