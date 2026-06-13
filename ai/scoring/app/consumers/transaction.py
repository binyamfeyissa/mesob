"""Consume TransactionPosted events to update feature store."""
import json
from app.infra.redis_store import get_features, set_features


def handle_transaction_posted(event: dict) -> None:
    """Update savings_velocity and transaction_frequency features."""
    user_id = event.get("user_id") or event.get("from_user_id")
    if not user_id:
        return

    amount_minor = int(event.get("amount_minor", 0))
    direction = event.get("direction", "")  # "C" credit / "D" debit

    features = get_features(user_id) or {}

    # transaction_frequency: rolling count (simple increment; TTL on the key handles windowing)
    features["transaction_frequency"] = int(features.get("transaction_frequency", 0)) + 1

    # savings_velocity: only positive credit events grow savings
    if direction == "C":
        features["savings_velocity_minor"] = (
            int(features.get("savings_velocity_minor", 0)) + amount_minor
        )

    set_features(user_id, features)
