from pydantic import BaseModel
from typing import Optional, List


class ScoreRequest(BaseModel):
    user_id: str
    force_recompute: bool = False


class FactorItem(BaseModel):
    feature: str
    contribution: float


class ScoreResponse(BaseModel):
    score_id: str
    score: int
    tier: str
    ceiling_minor: int
    model_ver: str
    source: str  # ML | RULES
    factors: List[FactorItem]


class ModelInfo(BaseModel):
    version: str
    status: str
    auc: Optional[float] = None
    fairness: Optional[float] = None
    promoted_at: Optional[str] = None


class PromoteRequest(BaseModel):
    canary_pct: int
    second_authoriser_id: str
