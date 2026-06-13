"""Unit tests for the rules scorecard — no external dependencies."""
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from app.domain.scorecard import ScorecardInput, run_scorecard
from app.domain.tiers import score_to_tier
from app.domain.fallback import compute_score, InsufficientHistoryError


def test_base_score():
    inp = ScorecardInput(
        months_active=6,
        iqub_cycles_completed=6,
        iqub_cycles_missed=0,
        avg_monthly_balance_minor=50000,
        prior_loan_repaid=False,
        prior_loan_defaulted=False,
    )
    result = run_scorecard(inp)
    assert 300 <= result.score <= 1000
    assert len(result.factors) > 0


def test_defaulted_lowers_score():
    base = ScorecardInput(
        months_active=6,
        iqub_cycles_completed=6,
        iqub_cycles_missed=0,
        avg_monthly_balance_minor=50000,
        prior_loan_repaid=False,
        prior_loan_defaulted=False,
    )
    with_default = ScorecardInput(
        months_active=6,
        iqub_cycles_completed=6,
        iqub_cycles_missed=0,
        avg_monthly_balance_minor=50000,
        prior_loan_repaid=False,
        prior_loan_defaulted=True,
    )
    assert run_scorecard(with_default).score < run_scorecard(base).score


def test_insufficient_history_raises():
    inp = ScorecardInput(
        months_active=0,
        iqub_cycles_completed=0,
        iqub_cycles_missed=0,
        avg_monthly_balance_minor=0,
        prior_loan_repaid=False,
        prior_loan_defaulted=False,
    )
    try:
        compute_score(inp, ml_available=False)
        assert False, "Expected InsufficientHistoryError"
    except InsufficientHistoryError:
        pass


def test_tier_mapping():
    # score_to_tier returns (tier_label, ceiling_minor)
    assert score_to_tier(300)[0] == "D"
    assert score_to_tier(500)[0] == "C"
    assert score_to_tier(600)[0] == "B"
    assert score_to_tier(750)[0] == "A"
    assert score_to_tier(900)[0] == "S"


def test_repaid_improves_score():
    no_history = ScorecardInput(
        months_active=6,
        iqub_cycles_completed=6,
        iqub_cycles_missed=0,
        avg_monthly_balance_minor=50000,
        prior_loan_repaid=False,
        prior_loan_defaulted=False,
    )
    with_repaid = ScorecardInput(
        months_active=6,
        iqub_cycles_completed=6,
        iqub_cycles_missed=0,
        avg_monthly_balance_minor=50000,
        prior_loan_repaid=True,
        prior_loan_defaulted=False,
    )
    assert run_scorecard(with_repaid).score > run_scorecard(no_history).score
