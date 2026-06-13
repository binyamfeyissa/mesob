"""Redis feature store for online scoring."""
import json
import os
from typing import Optional

import redis


def get_client() -> redis.Redis:
    url = os.getenv("MESOB_SCORING_REDIS_URL", "redis://localhost:6379/0")
    return redis.from_url(url, decode_responses=True)


def get_features(user_id: str) -> Optional[dict]:
    """Fetch user feature vector from Redis."""
    client = get_client()
    raw = client.get(f"features:{user_id}")
    if raw is None:
        return None
    return json.loads(raw)


def set_features(user_id: str, features: dict, ttl_seconds: int = 86400) -> None:
    """Write/update user feature vector in Redis."""
    client = get_client()
    client.setex(f"features:{user_id}", ttl_seconds, json.dumps(features))
