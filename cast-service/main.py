"""
Cast-to-Voice-Chat service
Receives stream URL from Go backend and plays it in a Telegram group Voice Chat.
"""

import os
import asyncio
from typing import Optional

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from pyrogram import Client
from pytgcalls import PyTgCalls
from pytgcalls.types.input_stream import AudioPiped
from pytgcalls.types.input_stream.quality import HighQualityAudio

load_dotenv()

API_ID = int(os.getenv("API_ID", "0"))
API_HASH = os.getenv("API_HASH", "")
SESSION_STRING = os.getenv("SESSION_STRING", "")  # preferred
# or use session file name
SESSION_NAME = os.getenv("SESSION_NAME", "music_cast")

app = FastAPI(title="Music Cast Service")

pyro: Optional[Client] = None
calls: Optional[PyTgCalls] = None


class CastBody(BaseModel):
    chat_id: int
    stream_url: str
    title: str = ""
    artist: str = ""
    track_id: str = ""


@app.on_event("startup")
async def startup():
    global pyro, calls
    if not API_ID or not API_HASH:
        print("[WARN] API_ID / API_HASH missing — cast service will not work until configured")
        return

    if SESSION_STRING:
        pyro = Client("music_cast", api_id=API_ID, api_hash=API_HASH, session_string=SESSION_STRING)
    else:
        pyro = Client(SESSION_NAME, api_id=API_ID, api_hash=API_HASH)

    await pyro.start()
    calls = PyTgCalls(pyro)
    await calls.start()
    print("[OK] Cast service ready")


@app.on_event("shutdown")
async def shutdown():
    if calls:
        await calls.stop()
    if pyro:
        await pyro.stop()


@app.get("/health")
async def health():
    return {"status": "ok", "ready": calls is not None}


@app.post("/cast")
async def cast(body: CastBody):
    if not calls or not pyro:
        raise HTTPException(503, "Cast service not configured (missing API_ID/API_HASH/session)")

    try:
        # Leave previous stream in this chat if any
        try:
            await calls.leave_group_call(body.chat_id)
        except Exception:
            pass

        await calls.join_group_call(
            body.chat_id,
            AudioPiped(
                body.stream_url,
                HighQualityAudio(),
            ),
        )
        return {
            "status": "playing",
            "chat_id": body.chat_id,
            "title": body.title,
            "artist": body.artist,
        }
    except Exception as e:
        raise HTTPException(500, f"cast failed: {str(e)}")


@app.post("/stop")
async def stop(chat_id: int):
    if not calls:
        raise HTTPException(503, "not ready")
    try:
        await calls.leave_group_call(chat_id)
        return {"status": "stopped"}
    except Exception as e:
        raise HTTPException(500, str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=4000, reload=True)
