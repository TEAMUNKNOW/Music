# Copyright (c) 2025-2026 TEAMUNKNOW
# Licensed under the MIT License.
# This file is part of Nova Music

import asyncio
import logging
import time
from logging.handlers import RotatingFileHandler

# Logging Configuration
logging.basicConfig(
    format="[%(asctime)s - %(levelname)s] - %(name)s: %(message)s",
    datefmt="%d-%b-%y %H:%M:%S",
    handlers=[
        RotatingFileHandler("log.txt", maxBytes=10485760, backupCount=5),
        logging.StreamHandler(),
    ],
    level=logging.INFO,
)

# Suppress noisy logs
logging.getLogger("httpx").setLevel(logging.ERROR)
logging.getLogger("ntgcalls").setLevel(logging.CRITICAL)
logging.getLogger("pymongo").setLevel(logging.ERROR)
logging.getLogger("pyrogram").setLevel(logging.ERROR)
logging.getLogger("pytgcalls").setLevel(logging.ERROR)

logger = logging.getLogger(__name__)

__version__ = "3.1.0"

# Config Initialization
from config import Config

config = Config()
config.check()

tasks = []
boot = time.time()

# Core Components
from anony.core.bot import Bot
app = Bot()

from anony.core.dir import ensure_dirs
ensure_dirs()

from anony.core.userbot import Userbot
userbot = Userbot()

from anony.core.mongo import MongoDB
db = MongoDB()

from anony.core.lang import Language
lang = Language()

from anony.core.telegram import Telegram
from anony.core.youtube import YouTube
from anony.core.soundcloud import SoundCloud
from anony.core.jiosaavn import JioSaavn
tg = Telegram()
yt = YouTube()
sc = SoundCloud()
js = JioSaavn()

from anony.helpers import Queue, Thumbnail
queue = Queue()
thumb = Thumbnail()

from anony.core.calls import TgCall
anon = TgCall()


async def stop() -> None:
    """Gracefully shutdown all running tasks and clients."""
    logger.info("Stopping bot services...")

    # Cancel all background tasks
    for task in tasks:
        task.cancel()
        try:
            await task
        except (asyncio.CancelledError, asyncio.exceptions.CancelledError):
            pass

    # Close core clients
    await app.exit()
    await userbot.exit()
    await db.close()
    await thumb.close()

    # Safely close YouTube session if active
    if getattr(yt, "api", None) and hasattr(yt.api, "session"):
        await yt.api.session.close()

    logger.info("Nova Music services stopped successfully.\n")
