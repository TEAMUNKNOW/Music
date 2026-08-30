# Copyright (c) 2025-2026 TEAMUNKNOW
# Licensed under the MIT License.
# This file is part of Nova Music

import asyncio
import importlib
import signal
from contextlib import suppress

from anony import anon, app, config, db, logger, stop, thumb, userbot, yt
from anony.plugins import all_modules


async def idle():
    """Wait for termination signals to gracefully shut down the bot."""
    loop = asyncio.get_running_loop()
    stop_event = asyncio.Event()

    for sig in (signal.SIGINT, signal.SIGTERM, signal.SIGABRT):
        with suppress(NotImplementedError, ValueError):
            loop.add_signal_handler(sig, stop_event.set)

    await stop_event.wait()


async def main():
    """Bootstraps all services, plugins, and handles runtime lifecycle."""
    await db.connect()
    await app.boot()
    await userbot.boot()
    await anon.boot()
    await thumb.start()

    # Load plugins
    for module in all_modules:
        importlib.import_module(f"anony.plugins.{module}")
    logger.info(f"Loaded {len(all_modules)} modules.")

    # YouTube setup & cookies
    if getattr(config, "COOKIES_URL", None):
        await yt.save_cookies(config.COOKIES_URL)

    if getattr(yt, "api", None) and hasattr(yt.api, "get_session"):
        await yt.api.get_session()

    # Sync sudoers & blacklisted users
    sudoers = await db.get_sudoers()
    app.sudoers.update(sudoers)
    app.bl_users.update(await db.get_blacklisted())
    logger.info(f"Loaded {len(app.sudoers)} sudo users.")

    # Keep bot alive until signaled to stop
    try:
        await idle()
    finally:
        await stop()


if __name__ == "__main__":
    try:
        asyncio.get_event_loop().run_until_complete(main())
    except KeyboardInterrupt:
        pass
