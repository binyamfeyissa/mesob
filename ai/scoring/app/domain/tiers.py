from dataclasses import dataclass
from typing import Tuple


@dataclass
class Tier:
    label: str
    min_score: int
    max_score: int
    ceiling_minor: int  # ETB cents


TIERS = [
    Tier("D", 0, 399, 0),
    Tier("C", 400, 549, 100_000_00),   # ETB 1,000
    Tier("B", 550, 699, 500_000_00),   # ETB 5,000
    Tier("A", 700, 849, 2_000_000_00), # ETB 20,000
    Tier("S", 850, 1000, 5_000_000_00), # ETB 50,000
]


def score_to_tier(score: int) -> Tuple[str, int]:
    """Returns (tier_label, ceiling_minor)."""
    for tier in reversed(TIERS):
        if score >= tier.min_score:
            return tier.label, tier.ceiling_minor
    return "D", 0
