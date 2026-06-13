"""Deterministic fraud rules: fast, explainable, no ML dependency."""
from dataclasses import dataclass, field
from typing import List


@dataclass
class RulesContext:
    user_id: str
    txn_type: str          # P2P, MERCHANT, BILL, CASH_IN, CASH_OUT
    amount_minor: int      # ETB cents — always int64, never float
    counterparty_id: str
    channel: str           # APP, USSD, AGENT
    # velocity features (from Redis feature store)
    count_1h: int = 0
    volume_1h_minor: int = 0
    count_24h: int = 0
    unique_counterparties_1h: int = 0
    agent_float_minor: int = 0
    agent_float_ceiling_minor: int = 0


@dataclass
class RulesResult:
    decision: str               # ALLOW | REVIEW | BLOCK
    rules_hit: List[str] = field(default_factory=list)
    risk_score: float = 0.0


# Structuring threshold: transactions near 10,000 ETB (1,000,000 minor) are suspicious
_STRUCTURING_THRESHOLD_MINOR = 1_000_000
_STRUCTURING_WINDOW = 0.85  # 85% of threshold


def run_rules(ctx: RulesContext) -> RulesResult:
    rules_hit = []
    risk_score = 0.0

    # Rule 1: High velocity — more than 10 txns in 1 hour
    if ctx.count_1h > 10:
        rules_hit.append("HIGH_VELOCITY_1H")
        risk_score += 0.4

    # Rule 2: High volume — more than 50k ETB in 1 hour
    if ctx.volume_1h_minor > 5_000_000:
        rules_hit.append("HIGH_VOLUME_1H")
        risk_score += 0.3

    # Rule 3: Daily velocity — more than 30 txns in 24 hours
    if ctx.count_24h > 30:
        rules_hit.append("HIGH_VELOCITY_24H")
        risk_score += 0.2

    # Rule 4: Structuring detection — amount near 10k ETB threshold
    if (_STRUCTURING_THRESHOLD_MINOR * _STRUCTURING_WINDOW
            <= ctx.amount_minor
            < _STRUCTURING_THRESHOLD_MINOR):
        rules_hit.append("STRUCTURING_NEAR_THRESHOLD")
        risk_score += 0.5

    # Rule 5: Fan-out — many unique counterparties in 1 hour (potential smurfing)
    if ctx.unique_counterparties_1h > 5:
        rules_hit.append("FAN_OUT_COUNTERPARTIES")
        risk_score += 0.3

    # Rule 6: Near-ceiling float (agent cash-in exceeds 90% of float ceiling)
    if (ctx.txn_type == "CASH_IN"
            and ctx.agent_float_ceiling_minor > 0
            and ctx.agent_float_minor >= ctx.agent_float_ceiling_minor * 0.9):
        rules_hit.append("NEAR_CEILING_FLOAT")
        risk_score += 0.2

    risk_score = min(1.0, risk_score)

    if risk_score >= 0.7:
        decision = "BLOCK"
    elif risk_score >= 0.3:
        decision = "REVIEW"
    else:
        decision = "ALLOW"

    return RulesResult(decision=decision, rules_hit=rules_hit, risk_score=risk_score)
