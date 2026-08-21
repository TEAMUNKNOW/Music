import pytest

from music.queue import QueueManager, Track


@pytest.mark.asyncio
async def test_queue_isolated_per_chat() -> None:
    queues = QueueManager()
    first = Track("one", "u1", "w1", 10, 1)
    second = Track("two", "u2", "w2", 20, 2)

    await queues.enqueue(100, first)
    await queues.enqueue(200, second)

    current_a, pending_a = await queues.snapshot(100)
    current_b, pending_b = await queues.snapshot(200)

    assert current_a is None
    assert current_b is None
    assert pending_a == [first]
    assert pending_b == [second]


@pytest.mark.asyncio
async def test_clear_removes_current_and_pending() -> None:
    queues = QueueManager()
    track = Track("one", "u1", "w1", 10, 1)
    await queues.enqueue(100, track)
    await queues.set_current(100, track)
    await queues.clear(100)

    current, pending = await queues.snapshot(100)
    assert current is None
    assert pending == []
