"""Unit tests for fraud rules — no external dependencies."""
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from app.domain.rules import RulesContext, run_rules


def make_ctx(**kwargs) -> RulesContext:
    defaults = dict(
        user_id="u1", txn_type="P2P", amount_minor=10000,
        counterparty_id="u2", channel="APP",
        count_1h=0, volume_1h_minor=0, count_24h=0,
        unique_counterparties_1h=0,
        agent_float_minor=0, agent_float_ceiling_minor=0,
    )
    defaults.update(kwargs)
    return RulesContext(**defaults)


def test_normal_transaction_allowed():
    result = run_rules(make_ctx(amount_minor=50000, count_1h=1))
    assert result.decision == "ALLOW"
    assert result.risk_score < 0.3


def test_high_velocity_triggers_review():
    result = run_rules(make_ctx(count_1h=15))
    assert result.decision in ("REVIEW", "BLOCK")
    assert "HIGH_VELOCITY_1H" in result.rules_hit


def test_structuring_near_threshold():
    # 890,000 minor = ~8,900 ETB — near 10,000 ETB threshold
    result = run_rules(make_ctx(amount_minor=890_000))
    assert "STRUCTURING_NEAR_THRESHOLD" in result.rules_hit


def test_high_risk_blocked():
    result = run_rules(make_ctx(
        count_1h=20,
        volume_1h_minor=8_000_000,
        amount_minor=890_000,
        unique_counterparties_1h=8,
    ))
    assert result.decision == "BLOCK"
    assert result.risk_score >= 0.7


def test_near_ceiling_float():
    result = run_rules(make_ctx(
        txn_type="CASH_IN",
        agent_float_minor=950_000,
        agent_float_ceiling_minor=1_000_000,
    ))
    assert "NEAR_CEILING_FLOAT" in result.rules_hit
