"""PostgreSQL persistence for credit scores."""
import os
import json
import uuid
from datetime import datetime, timezone
from typing import Optional

import psycopg2
import psycopg2.extras


def get_connection():
    url = os.getenv(
        "MESOB_SCORING_DB_URL",
        "postgresql://mesob:mesob@localhost:5432/mesob_scoring",
    )
    return psycopg2.connect(url, cursor_factory=psycopg2.extras.RealDictCursor)


def save_score(
    user_id: str,
    score: int,
    tier: str,
    ceiling_minor: int,
    model_ver: str,
    source: str,
    factors: list[dict],
    inputs_hash: bytes,
) -> str:
    """Persist a credit score record. Returns the new score_id."""
    score_id = str(uuid.uuid4())
    conn = get_connection()
    try:
        with conn, conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO credit_scores
                    (score_id, user_id, score, tier, ceiling_minor,
                     model_ver, source, factors, inputs_hash, created_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                """,
                (
                    score_id,
                    user_id,
                    score,
                    tier,
                    ceiling_minor,
                    model_ver,
                    source,
                    psycopg2.extras.Json(factors),
                    inputs_hash,
                    datetime.now(timezone.utc),
                ),
            )
    finally:
        conn.close()
    return score_id


def get_latest_score(user_id: str) -> Optional[dict]:
    """Fetch the most recent score for a user."""
    conn = get_connection()
    try:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT score_id, user_id, score, tier, ceiling_minor,
                       model_ver, source, factors, created_at
                FROM credit_scores
                WHERE user_id = %s
                ORDER BY created_at DESC
                LIMIT 1
                """,
                (user_id,),
            )
            row = cur.fetchone()
            if row is None:
                return None
            return {
                "score_id": str(row["score_id"]),
                "user_id": str(row["user_id"]),
                "score": row["score"],
                "tier": row["tier"],
                "ceiling_minor": row["ceiling_minor"],
                "model_ver": row["model_ver"],
                "source": row["source"],
                "factors": row["factors"] if isinstance(row["factors"], list) else json.loads(row["factors"] or "[]"),
                "created_at": row["created_at"].isoformat() if row["created_at"] else None,
            }
    finally:
        conn.close()
