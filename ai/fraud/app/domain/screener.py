"""Hybrid screener: deterministic rules first, then anomaly detection."""
from dataclasses import dataclass, field
from typing import List

from app.domain.rules import RulesContext, RulesResult, run_rules
from app.domain.anomaly import detector


@dataclass
class ScreenResult:
    decision: str          # ALLOW | REVIEW | BLOCK
    risk_score: float
    rules_hit: List[str] = field(default_factory=list)


def screen(ctx: RulesContext) -> ScreenResult:
    """
    Two-stage screening:
    1. Deterministic rules (fast, explainable) — BLOCK is final
    2. Anomaly detection (novel patterns) — can escalate ALLOW -> REVIEW
    """
    rules_result: RulesResult = run_rules(ctx)

    # Rules BLOCK is final — don't need anomaly score
    if rules_result.decision == "BLOCK":
        return ScreenResult(
            decision="BLOCK",
            risk_score=rules_result.risk_score,
            rules_hit=rules_result.rules_hit,
        )

    # Attempt anomaly scoring on top of rules
    features = {
        "amount_minor": ctx.amount_minor,
        "count_1h": ctx.count_1h,
        "volume_1h_minor": ctx.volume_1h_minor,
        "count_24h": ctx.count_24h,
        "unique_counterparties_1h": ctx.unique_counterparties_1h,
    }
    anomaly_score = detector.score(features)

    risk_score = rules_result.risk_score
    if anomaly_score is not None:
        # Blend: 60% rules, 40% anomaly
        risk_score = 0.6 * rules_result.risk_score + 0.4 * anomaly_score

    if risk_score >= 0.7:
        decision = "BLOCK"
    elif risk_score >= 0.3:
        decision = "REVIEW"
    else:
        decision = "ALLOW"

    return ScreenResult(
        decision=decision,
        risk_score=risk_score,
        rules_hit=rules_result.rules_hit,
    )
