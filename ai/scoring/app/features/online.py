"""
Online feature store reads — Redis hash per user.
Keys: mesob:features:{user_id}:{feature_name}
"""
import json
from typing import Optional, Dict, Any


class FeatureStore:
    def __init__(self, redis_client):
        self.redis = redis_client

    def get_features(self, user_id: str) -> Optional[Dict[str, Any]]:
        key = f"mesob:features:{user_id}"
        raw = self.redis.hgetall(key)
        if not raw:
            return None
        return {k.decode(): json.loads(v) for k, v in raw.items()}

    def update_feature(self, user_id: str, feature: str, value: Any) -> None:
        key = f"mesob:features:{user_id}"
        self.redis.hset(key, feature, json.dumps(value))
