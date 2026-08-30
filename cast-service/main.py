"""
Optional Cast-to-Voice-Chat service.

The service always starts normally. VC casting remains disabled unless
API_ID, API_HASH and SESSION_STRING are explicitly configured.
"""

import os

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

load_dotenv()

API_ID = int(os.getenv("API_ID", "0") or "0")
API_HASH = os.getenv("API_HASH", "") or ""
SESSION_STRING = os.getenv("SESSION_STRING", "") or ""
SESSION_NAME = os.getenv("SESSION_NAME", "music_cast") or "music_cast"

app = FastAPI(title="Music Cast Service (Optional)")

pyro = None
calls = None
CAST_ENABLED = False


class CastBody(BaseModel):
    chat_id: int
    stream_url: str
    title: str = ""
    artist: str = ""
    track_id: str = ""


def _has_credentials() -> bool:
    return bool(API_ID and API_HASH and SESSION_STRING)


@app.on_event("startup")
async def startup():
    global pyro, calls, CAST_ENABLED

    # No credentials = service stays healthy, but VC casting stays OFF.
    if not _has_credentials():
        print("[INFO] VC Cast disabled — API_ID/API_HASH/SESSION_STRING not configured")
        CAST_ENABLED = False
        return

    try:
        from pyrogram import Client
        from pytgcalls import PyTgCalls

        pyro = Client(
            "music_cast",
            api_id=API_ID,
            api_hash=API_HASH,
            session_string=SESSION_STRING,
        )
        await pyro.start()
        calls = PyTgCalls(pyro)
        await calls.start()
        CAST_ENABLED = True
        print("[OK] VC Cast enabled — userbot ready")
    except Exception as e:
        print(f"[WARN] Failed to start VC Cast: {e}")
        CAST_ENABLED = False
        pyro = None
        calls = None


@app.on_event("shutdown")
async def shutdown():
    global calls, pyro
    if calls:
        try:
            await calls.stop()
        except Exception:
            pass
    if pyro:
        try:
            await pyro.stop()
        except Exception:
            pass


@app.get("/health")
async def health():
    return {
        "status": "ok",
        "cast_enabled": CAST_ENABLED,
        "ready": CAST_ENABLED,
    }


@app.post("/cast")
async def cast(body: CastBody):
    if not CAST_ENABLED or not calls:
        raise HTTPException(
            status_code=503,
            detail="VC Cast is disabled. Configure API_ID, API_HASH and SESSION_STRING to enable it.",
        )

    try:
        from pytgcalls.types.input_stream import AudioPiped
        from pytgcalls.types.input_stream.quality import HighQualityAudio

        try:
            await calls.leave_group_call(body.chat_id)
        except Exception:
            pass

        await calls.join_group_call(
            body.chat_id,
            AudioPiped(body.stream_url, HighQualityAudio()),
        )
        return {
            "status": "playing",
            "chat_id": body.chat_id,
            "title": body.title,
            "artist": body.artist,
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"cast failed: {str(e)}")


@app.post("/stop")
async def stop(chat_id: int):
    if not CAST_ENABLED or not calls:
        raise HTTPException(status_code=503, detail="VC Cast is disabled")
    try:
        await calls.leave_group_call(chat_id)
        return {"status": "stopped"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=4000, reload=False)
