"""Consume IqubContribution events to update iqub_punctuality feature."""
from app.infra.redis_store import get_features, set_features


def _update_punctuality(user_id: str, on_time: bool) -> None:
    features = get_features(user_id) or {}
    total = int(features.get("iqub_total", 0)) + 1
    on_time_count = int(features.get("iqub_on_time", 0)) + (1 if on_time else 0)
    features["iqub_total"] = total
    features["iqub_on_time"] = on_time_count
    features["iqub_punctuality"] = round(on_time_count / total, 4) if total > 0 else 0.0
    set_features(user_id, features)


def handle_iqub_contribution_recorded(event: dict) -> None:
    """Increment on-time contribution count and recalculate punctuality ratio."""
    user_id = event.get("user_id")
    if not user_id:
        return
    _update_punctuality(user_id, on_time=True)


def handle_iqub_contribution_missed(event: dict) -> None:
    """Increment missed contribution count and recalculate punctuality ratio."""
    user_id = event.get("user_id")
    if not user_id:
        return
    _update_punctuality(user_id, on_time=False)
