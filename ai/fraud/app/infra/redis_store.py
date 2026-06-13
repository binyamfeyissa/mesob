"""Redis velocity feature store for fraud rules."""
import json
import os
from typing import Optional

import redis


def get_client() -> redis.Redis:
    url = os.getenv("MESOB_FRAUD_REDIS_URL", "redis://localhost:6379/1")
    return redis.from_url(url, decode_responses=True)


def get_velocity(user_id: str) -> dict:
    """
    Fetch pre-computed velocity counters for a user.
    Returns a dict with count_1h, volume_1h_minor, count_24h,
    unique_counterparties_1h, agent_float_minor, agent_float_ceiling_minor.
    """
    client = get_client()
    pipe = client.pipeline(transaction=False)
    pipe.get(f"fraud:count_1h:{user_id}")
    pipe.get(f"fraud:volume_1h:{user_id}")
    pipe.get(f"fraud:count_24h:{user_id}")
    pipe.scard(f"fraud:cparties_1h:{user_id}")
    pipe.get(f"fraud:agent_float:{user_id}")
    pipe.get(f"fraud:agent_float_ceil:{user_id}")
    results = pipe.execute()

    return {
        "count_1h": int(results[0] or 0),
        "volume_1h_minor": int(results[1] or 0),
        "count_24h": int(results[2] or 0),
        "unique_counterparties_1h": int(results[3] or 0),
        "agent_float_minor": int(results[4] or 0),
        "agent_float_ceiling_minor": int(results[5] or 0),
    }


def increment_velocity(user_id: str, amount_minor: int, counterparty_id: str) -> None:
    """
    Increment rolling counters after a transaction is processed.
    Called by Kafka consumer on TransactionPosted events.
    """
    client = get_client()
    pipe = client.pipeline(transaction=False)

    # 1-hour sliding counters
    pipe.incr(f"fraud:count_1h:{user_id}")
    pipe.expire(f"fraud:count_1h:{user_id}", 3600)

    pipe.incrby(f"fraud:volume_1h:{user_id}", amount_minor)
    pipe.expire(f"fraud:volume_1h:{user_id}", 3600)

    # unique counterparties in last hour (set expires on the set key)
    pipe.sadd(f"fraud:cparties_1h:{user_id}", counterparty_id)
    pipe.expire(f"fraud:cparties_1h:{user_id}", 3600)

    # 24-hour counter
    pipe.incr(f"fraud:count_24h:{user_id}")
    pipe.expire(f"fraud:count_24h:{user_id}", 86400)

    pipe.execute()
