import os
from dataclasses import dataclass

from dotenv import load_dotenv

load_dotenv()


@dataclass(frozen=True)
class Config:
    api_id: int
    api_hash: str
    bot_token: str

    @classmethod
    def from_env(cls) -> "Config":
        missing = [key for key in ("API_ID", "API_HASH", "BOT_TOKEN") if not os.getenv(key)]
        if missing:
            raise RuntimeError(f"Missing environment variables: {', '.join(missing)}")
        try:
            api_id = int(os.environ["API_ID"])
        except ValueError as exc:
            raise RuntimeError("API_ID must be an integer") from exc
        return cls(api_id=api_id, api_hash=os.environ["API_HASH"], bot_token=os.environ["BOT_TOKEN"])


config = Config.from_env()
