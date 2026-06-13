"""PostgreSQL persistence for fraud alerts."""
import os
import uuid
from datetime import datetime, timezone
from typing import Optional

import psycopg2
import psycopg2.extras


def get_connection():
    url = os.getenv(
        "MESOB_FRAUD_DB_URL",
        "postgresql://mesob:mesob@localhost:5432/mesob_fraud",
    )
    return psycopg2.connect(url, cursor_factory=psycopg2.extras.RealDictCursor)


def save_alert(
    transaction_id: str,
    user_id: str,
    risk_score: float,
    rules_hit: list[str],
    decision: str,
) -> str:
    """Persist a BLOCK or REVIEW fraud alert. Returns alert_id."""
    alert_id = str(uuid.uuid4())
    severity = "HIGH" if risk_score >= 0.7 else "MEDIUM"
    primary_rule = rules_hit[0] if rules_hit else "UNKNOWN"
    conn = get_connection()
    try:
        with conn, conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO fraud_alerts
                    (id, transaction_id, user_id, risk_score, rules_hit,
                     decision, severity, primary_rule, status, created_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, 'OPEN', %s)
                """,
                (
                    alert_id,
                    transaction_id,
                    user_id,
                    risk_score,
                    psycopg2.extras.Json(rules_hit),
                    decision,
                    severity,
                    primary_rule,
                    datetime.now(timezone.utc),
                ),
            )
    finally:
        conn.close()
    return alert_id


def get_open_alerts(limit: int = 50, cursor: Optional[str] = None) -> list[dict]:
    """Return open fraud alerts for admin review, newest first."""
    conn = get_connection()
    try:
        with conn.cursor() as cur:
            if cursor:
                cur.execute(
                    """
                    SELECT id AS alert_id, severity, primary_rule AS rule,
                           transaction_id, status, created_at
                    FROM fraud_alerts
                    WHERE status = 'OPEN' AND created_at < (
                        SELECT created_at FROM fraud_alerts WHERE id = %s
                    )
                    ORDER BY created_at DESC
                    LIMIT %s
                    """,
                    (cursor, limit),
                )
            else:
                cur.execute(
                    """
                    SELECT id AS alert_id, severity, primary_rule AS rule,
                           transaction_id, status, created_at
                    FROM fraud_alerts
                    WHERE status = 'OPEN'
                    ORDER BY created_at DESC
                    LIMIT %s
                    """,
                    (limit,),
                )
            rows = cur.fetchall()
            return [
                {
                    "alert_id": str(r["alert_id"]),
                    "severity": r["severity"],
                    "rule": r["rule"],
                    "transaction_id": str(r["transaction_id"]),
                    "status": r["status"],
                    "created_at": r["created_at"].isoformat() if r["created_at"] else None,
                }
                for r in (rows or [])
            ]
    finally:
        conn.close()


def update_alert_disposition(
    alert_id: str,
    disposition: str,
    note: str,
    actor_id: str,
) -> None:
    """
    Update alert status after admin disposition: CLEAR | CONFIRM | ESCALATE_SAR.
    Writes to audit_log as well.
    """
    status_map = {
        "CLEAR": "CLEARED",
        "CONFIRM": "CONFIRMED",
        "ESCALATE_SAR": "ESCALATED_SAR",
    }
    new_status = status_map.get(disposition, disposition)
    audit_id = str(uuid.uuid4())
    now = datetime.now(timezone.utc)

    conn = get_connection()
    try:
        with conn, conn.cursor() as cur:
            cur.execute(
                """
                UPDATE fraud_alerts
                SET status = %s,
                    disposition_note = %s,
                    resolved_by = %s,
                    resolved_at = %s
                WHERE id = %s
                """,
                (new_status, note, actor_id, now, alert_id),
            )
            cur.execute(
                """
                INSERT INTO fraud_audit_log
                    (id, alert_id, actor_id, action, note, created_at)
                VALUES (%s, %s, %s, %s, %s, %s)
                """,
                (audit_id, alert_id, actor_id, f"DISPOSITION_{disposition}", note, now),
            )
    finally:
        conn.close()
