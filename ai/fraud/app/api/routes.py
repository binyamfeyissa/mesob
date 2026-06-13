from fastapi import APIRouter, HTTPException, Header
from typing import Optional

from app.api.models import ScreenRequest, ScreenResponse, AlertRecord, DispositionRequest
from app.domain.rules import RulesContext
from app.domain.screener import screen
from app.infra import db, redis_store

router = APIRouter()


@router.post("/fraud/screen", response_model=ScreenResponse)
def screen_transaction(req: ScreenRequest) -> ScreenResponse:
    """
    Synchronous fraud screen on the money path.
    Fails closed: caller must decline if this endpoint is unavailable.
    BLOCK or REVIEW decision triggers alert persistence.
    """
    velocity = redis_store.get_velocity(req.user_id)

    ctx = RulesContext(
        user_id=req.user_id,
        txn_type=req.type,
        amount_minor=req.amount_minor,
        counterparty_id=req.counterparty,
        channel=req.channel,
        count_1h=velocity["count_1h"],
        volume_1h_minor=velocity["volume_1h_minor"],
        count_24h=velocity["count_24h"],
        unique_counterparties_1h=velocity["unique_counterparties_1h"],
        agent_float_minor=velocity["agent_float_minor"],
        agent_float_ceiling_minor=velocity["agent_float_ceiling_minor"],
    )
    result = screen(ctx)

    if result.decision in ("BLOCK", "REVIEW"):
        try:
            db.save_alert(
                transaction_id=req.counterparty,  # counterparty carries txn_id on this path
                user_id=req.user_id,
                risk_score=result.risk_score,
                rules_hit=result.rules_hit,
                decision=result.decision,
            )
        except Exception:
            pass  # alert persistence must not fail the payment path

    return ScreenResponse(
        decision=result.decision,
        risk_score=result.risk_score,
        rules_hit=result.rules_hit,
    )


@router.get("/fraud/alerts", response_model=list[AlertRecord])
def list_alerts(limit: int = 50, cursor: Optional[str] = None) -> list[AlertRecord]:
    """List open fraud alerts (SUPER_ADMIN)."""
    try:
        rows = db.get_open_alerts(limit=limit, cursor=cursor)
    except Exception as exc:
        raise HTTPException(status_code=503, detail={"code": "DB_UNAVAILABLE", "message": str(exc)})
    return [
        AlertRecord(
            alert_id=r["alert_id"],
            severity=r["severity"],
            rule=r["rule"],
            transaction_id=r["transaction_id"],
            status=r["status"],
        )
        for r in rows
    ]


@router.post("/fraud/alerts/{alert_id}/disposition")
def disposition_alert(
    alert_id: str,
    req: DispositionRequest,
    x_user_id: Optional[str] = Header(default=None, alias="X-User-ID"),
):
    """Disposition a fraud alert: CLEAR | CONFIRM | ESCALATE_SAR (SUPER_ADMIN, audited)."""
    valid_dispositions = {"CLEAR", "CONFIRM", "ESCALATE_SAR"}
    if req.disposition not in valid_dispositions:
        raise HTTPException(
            status_code=422,
            detail={"code": "INVALID_DISPOSITION", "message": f"Must be one of {valid_dispositions}"},
        )
    actor_id = x_user_id or "unknown"
    try:
        db.update_alert_disposition(
            alert_id=alert_id,
            disposition=req.disposition,
            note=req.note,
            actor_id=actor_id,
        )
    except Exception as exc:
        raise HTTPException(status_code=503, detail={"code": "DB_UNAVAILABLE", "message": str(exc)})
    return {"alert_id": alert_id, "disposition": req.disposition, "status": "processed"}
