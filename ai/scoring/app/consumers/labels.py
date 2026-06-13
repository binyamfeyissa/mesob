"""Consume loan outcome events to store training labels in Redis + invalidate score cache."""
import os
import psycopg2


def _get_conn():
    url = os.getenv("MESOB_SCORING_DB_URL", "postgresql://mesob:mesob@localhost:5432/mesob_scoring")
    return psycopg2.connect(url)


def _record_label(user_id: str, label: str) -> None:
    """Write a repayment label row; non-fatal if DB unavailable."""
    try:
        conn = _get_conn()
        with conn, conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO scoring_labels (user_id, label, created_at)
                VALUES (%s, %s, NOW())
                ON CONFLICT DO NOTHING
                """,
                (user_id, label),
            )
    except Exception:
        pass


def handle_loan_repaid(event: dict) -> None:
    """Record positive repayment label for model training."""
    user_id = event.get("user_id")
    if not user_id:
        return

    from app.infra.redis_store import get_features, set_features
    features = get_features(user_id) or {}
    features["prior_loan_repaid"] = int(features.get("prior_loan_repaid", 0)) + 1
    set_features(user_id, features)

    _record_label(user_id, "REPAID")


def handle_loan_defaulted(event: dict) -> None:
    """Record default label for model training."""
    user_id = event.get("user_id")
    if not user_id:
        return

    from app.infra.redis_store import get_features, set_features
    features = get_features(user_id) or {}
    features["prior_loan_defaulted"] = int(features.get("prior_loan_defaulted", 0)) + 1
    set_features(user_id, features)

    _record_label(user_id, "DEFAULTED")
