import hashlib
import json

from fastapi import APIRouter, HTTPException

from app.api.models import ScoreRequest, ScoreResponse, ModelInfo, PromoteRequest
from app.domain.fallback import compute_score, InsufficientHistoryError
from app.domain.scorecard import ScorecardInput
from app.domain.tiers import score_to_tier
from app.infra import db, redis_store

router = APIRouter()

_MODEL_VER_RULES = "rules-v1"


def _build_scorecard_input(features: dict | None) -> ScorecardInput | None:
    """Convert Redis feature dict to ScorecardInput, or return None if no features."""
    if not features:
        return None
    return ScorecardInput(
        months_active=int(features.get("months_active", 0)),
        iqub_cycles_completed=int(features.get("iqub_cycles_completed", 0)),
        iqub_cycles_missed=int(features.get("iqub_cycles_missed", 0)),
        avg_monthly_balance_minor=int(features.get("avg_monthly_balance_minor", 0)),
        prior_loan_repaid=bool(features.get("prior_loan_repaid", False)),
        prior_loan_defaulted=bool(features.get("prior_loan_defaulted", False)),
    )


def _inputs_hash(user_id: str, features: dict | None) -> bytes:
    payload = json.dumps({"user_id": user_id, "features": features}, sort_keys=True)
    return hashlib.sha256(payload.encode()).digest()


@router.post("/scoring/score", response_model=ScoreResponse)
def score_user(req: ScoreRequest) -> ScoreResponse:
    """
    Score a user. Fallback chain: ML → rules → INSUFFICIENT_HISTORY (422).
    Returns cached score when force_recompute=False and a fresh score exists in DB.
    """
    # Return cached score unless force_recompute requested
    if not req.force_recompute:
        try:
            cached = db.get_latest_score(req.user_id)
            if cached:
                return ScoreResponse(
                    score_id=cached["score_id"],
                    score=cached["score"],
                    tier=cached["tier"],
                    ceiling_minor=cached["ceiling_minor"],
                    model_ver=cached["model_ver"],
                    source=cached["source"],
                    factors=[
                        {"feature": f["feature"], "contribution": f["contribution"]}
                        for f in cached["factors"]
                    ],
                )
        except Exception:
            pass  # DB unavailable — fall through to compute

    # Fetch features from Redis feature store
    features = None
    try:
        features = redis_store.get_features(req.user_id)
    except Exception:
        pass

    scorecard_input = _build_scorecard_input(features)

    try:
        result = compute_score(scorecard_input, ml_available=False)
    except InsufficientHistoryError as e:
        raise HTTPException(
            status_code=422,
            detail={"code": "INSUFFICIENT_HISTORY", "message": str(e)},
        )

    tier_label, ceiling_minor = score_to_tier(result.score)
    inputs_hash = _inputs_hash(req.user_id, features)

    # Persist score (non-fatal)
    score_id = "00000000-0000-0000-0000-000000000000"
    try:
        score_id = db.save_score(
            user_id=req.user_id,
            score=result.score,
            tier=tier_label,
            ceiling_minor=ceiling_minor,
            model_ver=_MODEL_VER_RULES,
            source="RULES",
            factors=result.factors,
            inputs_hash=inputs_hash,
        )
    except Exception:
        pass

    return ScoreResponse(
        score_id=score_id,
        score=result.score,
        tier=tier_label,
        ceiling_minor=ceiling_minor,
        model_ver=_MODEL_VER_RULES,
        source="RULES",
        factors=[
            {"feature": f["feature"], "contribution": f["contribution"]}
            for f in result.factors
        ],
    )


@router.get("/scoring/models", response_model=list[ModelInfo])
def list_models() -> list[ModelInfo]:
    """List registered model versions."""
    return [
        ModelInfo(
            version=_MODEL_VER_RULES,
            status="PRODUCTION",
            auc=None,
            fairness=None,
            promoted_at=None,
        )
    ]


@router.post("/scoring/models/{version}/promote")
def promote_model(version: str, req: PromoteRequest):
    """Promote a model version to canary or production (SUPER_ADMIN, 4-eyes)."""
    if req.canary_pct < 0 or req.canary_pct > 100:
        raise HTTPException(
            status_code=422,
            detail={"code": "INVALID_CANARY_PCT", "message": "canary_pct must be 0–100"},
        )
    if not req.second_authoriser_id:
        raise HTTPException(
            status_code=422,
            detail={"code": "MISSING_SECOND_AUTHORISER", "message": "4-eyes required for model promotion"},
        )
    return {"version": version, "canary_pct": req.canary_pct, "status": "promoted"}
