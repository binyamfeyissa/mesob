"""
Cold-start rules scorecard — transparent, auditable, no ML required.
Used when: new user, insufficient history, ML model unavailable.
"""
from dataclasses import dataclass, field
from typing import Dict, List


@dataclass
class ScorecardInput:
    iqub_cycles_completed: int = 0
    iqub_cycles_missed: int = 0
    months_active: int = 0
    avg_monthly_balance_minor: int = 0
    prior_loan_repaid: bool = False
    prior_loan_defaulted: bool = False


@dataclass
class ScorecardResult:
    score: int
    factors: List[Dict]


def run_scorecard(inp: ScorecardInput) -> ScorecardResult:
    """
    Deterministic scorecard. Each rule contributes a fixed number of points.
    Total possible: 1000.
    """
    score = 300  # base score for any registered user
    factors = []

    # Iqub punctuality (max +200)
    total_cycles = inp.iqub_cycles_completed + inp.iqub_cycles_missed
    if total_cycles > 0:
        punctuality = inp.iqub_cycles_completed / total_cycles
        pts = int(punctuality * 200)
        score += pts
        factors.append({"feature": "iqub_punctuality", "contribution": pts / 1000.0})

    # Account tenure (max +100)
    tenure_pts = min(inp.months_active * 5, 100)
    score += tenure_pts
    factors.append({"feature": "months_active", "contribution": tenure_pts / 1000.0})

    # Savings velocity (max +150)
    if inp.avg_monthly_balance_minor >= 500_00:  # ETB 500
        balance_pts = min(int(inp.avg_monthly_balance_minor / 500_00) * 10, 150)
        score += balance_pts
        factors.append({"feature": "savings_velocity", "contribution": balance_pts / 1000.0})

    # Prior loan history (+200 / -300)
    if inp.prior_loan_repaid:
        score += 200
        factors.append({"feature": "prior_loan_repaid", "contribution": 0.2})
    if inp.prior_loan_defaulted:
        score -= 300
        factors.append({"feature": "prior_loan_defaulted", "contribution": -0.3})

    score = max(0, min(1000, score))
    return ScorecardResult(score=score, factors=factors)
