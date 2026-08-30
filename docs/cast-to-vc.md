# Cast to Group Voice Chat (Optional)

This feature is **completely optional**.

- Agar `SESSION_STRING` (ya session file) + `API_ID` / `API_HASH` **nahi** diye → Cast disabled
- Frontend me **Cast button hide** rehta hai
- Backend `/api/cast` 503 return karta hai with clear message
- Core music streaming (Mini App playback) bina userbot ke perfectly kaam karta hai

## Enable karne ke liye

1. https://my.telegram.org se `API_ID` + `API_HASH` lo
2. Ek baar session string generate karo (Pyrogram string session)
3. `cast-service` me set karo:

```env
API_ID=12345678
API_HASH=your_api_hash
SESSION_STRING=1BVtsO...   # preferred
```

4. Cast service start karo — ab `/health` me `cast_enabled: true` aayega
5. Frontend automatically Cast button dikhayega

## Flow (when enabled)

```
Mini App (group se open)
   → POST /api/cast { chat_id, track_id }
   → Backend resolves stream URL
   → cast-service userbot joins group VC
   → Audio plays in Voice Chat
```

## Without session

- Personal Mode (Mini App streaming) 100% works
- Search, Player, Waveform, Lyrics, Sync Rooms — sab chalega
- Sirf "Cast to Group Voice Chat" feature absent rahega
