"""
Fallback chain: ML → rules scorecard → INSUFFICIENT_HISTORY.
Never guesses. Never returns a score when there is not enough data.
"""
from typing import Optional
from .scorecard import ScorecardInput, ScorecardResult, run_scorecard


MIN_MONTHS_ACTIVE = 1  # minimum history required for any score


class InsufficientHistoryError(Exception):
    """Raised when user has insufficient transaction history."""
    pass


class ScoringDeferredError(Exception):
    """Raised when ML service is unavailable and rules cannot run."""
    pass


def compute_score(
    scorecard_input: Optional[ScorecardInput],
    ml_available: bool,
    ml_score: Optional[int] = None,
    ml_factors=None,
) -> ScorecardResult:
    """
    Fallback chain:
    1. If ML available and returned a score → use ML
    2. If ML unavailable → use rules scorecard (source: RULES)
    3. If insufficient history for rules → raise InsufficientHistoryError
    """
    if ml_available and ml_score is not None:
        return ScorecardResult(score=ml_score, factors=ml_factors or [])

    if scorecard_input is None:
        raise InsufficientHistoryError("No feature data available for this user")

    if scorecard_input.months_active < MIN_MONTHS_ACTIVE:
        raise InsufficientHistoryError(
            f"User has only {scorecard_input.months_active} months of history; "
            f"minimum required: {MIN_MONTHS_ACTIVE}"
        )

    return run_scorecard(scorecard_input)
