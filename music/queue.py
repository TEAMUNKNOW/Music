from __future__ import annotations

import asyncio
from collections import deque
from dataclasses import dataclass
from typing import Deque


@dataclass(frozen=True, slots=True)
class Track:
    title: str
    stream_url: str
    webpage_url: str
    duration: int | None
    requester_id: int


@dataclass(slots=True)
class QueueState:
    current: Track | None = None
    items: Deque[Track] | None = None
    lock: asyncio.Lock | None = None
    manual_action_at: float = 0.0

    def __post_init__(self) -> None:
        if self.items is None:
            self.items = deque()
        if self.lock is None:
            self.lock = asyncio.Lock()


class QueueManager:
    def __init__(self) -> None:
        self._states: dict[int, QueueState] = {}
        self._states_lock = asyncio.Lock()

    async def state(self, chat_id: int) -> QueueState:
        async with self._states_lock:
            return self._states.setdefault(chat_id, QueueState())

    async def enqueue(self, chat_id: int, track: Track) -> int:
        state = await self.state(chat_id)
        assert state.items is not None
        state.items.append(track)
        return len(state.items)

    async def set_current(self, chat_id: int, track: Track | None) -> None:
        state = await self.state(chat_id)
        state.current = track

    async def pop_next(self, chat_id: int) -> Track | None:
        state = await self.state(chat_id)
        assert state.items is not None
        return state.items.popleft() if state.items else None

    async def clear(self, chat_id: int) -> None:
        state = await self.state(chat_id)
        assert state.items is not None
        state.items.clear()
        state.current = None

    async def snapshot(self, chat_id: int) -> tuple[Track | None, list[Track]]:
        state = await self.state(chat_id)
        assert state.items is not None
        return state.current, list(state.items)

    async def mark_manual_action(self, chat_id: int, when: float) -> None:
        state = await self.state(chat_id)
        state.manual_action_at = when

    async def was_recent_manual_action(self, chat_id: int, now: float, window: float = 1.5) -> bool:
        state = await self.state(chat_id)
        return now - state.manual_action_at < window
