"""
Batch feature computation — runs offline via Spark or scheduled SQL job.
Reads from postgres, writes to Redis feature store.
"""


def compute_iqub_features(db_conn, user_id: str) -> dict:
    """TODO: implement — query iqub_memberships and iqub_cycles."""
    return {
        "iqub_cycles_completed": 0,
        "iqub_cycles_missed": 0,
    }


def compute_savings_features(db_conn, user_id: str) -> dict:
    """TODO: implement — query ledger_entries for user wallet account."""
    return {
        "avg_monthly_balance_minor": 0,
        "months_active": 0,
    }


def compute_loan_features(db_conn, user_id: str) -> dict:
    """TODO: implement — query loans and repayments."""
    return {
        "prior_loan_repaid": False,
        "prior_loan_defaulted": False,
    }
