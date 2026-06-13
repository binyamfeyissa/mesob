"""
Feature family definitions.
All features are computed from observable financial behaviour — no ethnicity, religion, or protected attributes.
"""

FEATURE_FAMILIES = {
    "iqub_punctuality": "Ratio of on-time Iqub contributions to total cycle obligations",
    "savings_velocity": "Average monthly savings balance growth rate",
    "transaction_frequency": "Number of transactions per month (30-day rolling)",
    "repayment_history": "Loan repayment rate on prior loans",
    "account_tenure": "Months since first transaction",
    "float_utilisation": "Agent float utilisation ratio (agents only)",
    "iddir_coverage": "Continuous Iddir premium payment streak",
}

ML_FEATURE_NAMES = list(FEATURE_FAMILIES.keys())
