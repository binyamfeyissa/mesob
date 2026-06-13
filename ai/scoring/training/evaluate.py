"""Evaluate model: AUC + disparate-impact fairness gates.

Fairness constraint: no protected attributes (gender, ethnicity) in features.
Disparate-impact is computed on loan decision outcomes by self-reported gender
if that column is present in the test data. Gate passes if DI ratio >= 0.8.
"""
import json
import os
import pickle

import numpy as np


FEATURE_COLS = [
    "iqub_punctuality",
    "savings_velocity_minor",
    "transaction_frequency",
    "prior_loan_repaid",
    "prior_loan_defaulted",
    "account_tenure_months",
]

AUC_THRESHOLD = 0.65
DI_THRESHOLD = 0.80


def evaluate(model_path: str, test_data_path: str) -> dict:
    """Evaluate a trained model against test data. Returns metrics dict."""
    try:
        import pandas as pd
        from sklearn.metrics import roc_auc_score
    except ImportError:
        return {"auc": 0.0, "disparate_impact": 1.0, "passed": False, "error": "missing sklearn/pandas"}

    with open(model_path, "rb") as f:
        model = pickle.load(f)

    if test_data_path.endswith(".csv"):
        import pandas as pd
        df = pd.read_csv(test_data_path)
    else:
        import pandas as pd
        with open(test_data_path) as fh:
            rows = [json.loads(line) for line in fh if line.strip()]
        df = pd.DataFrame(rows)

    available = [c for c in FEATURE_COLS if c in df.columns]
    X = df[available].fillna(0).values
    y_true = df["repaid"].values

    y_prob = model.predict_proba(X)[:, 1]
    auc = float(roc_auc_score(y_true, y_prob))

    # Disparate-impact: approval rate ratio (minority group / majority group).
    di = 1.0
    if "gender" in df.columns:
        y_pred = (y_prob >= 0.5).astype(int)
        groups = df["gender"].unique()
        rates = {g: y_pred[df["gender"] == g].mean() for g in groups if len(df[df["gender"] == g]) > 10}
        if len(rates) >= 2:
            sorted_rates = sorted(rates.values())
            di = float(sorted_rates[0] / sorted_rates[-1]) if sorted_rates[-1] > 0 else 1.0

    passed = auc >= AUC_THRESHOLD and di >= DI_THRESHOLD

    print(f"AUC: {auc:.4f} (threshold={AUC_THRESHOLD}) {'✓' if auc >= AUC_THRESHOLD else '✗'}")
    print(f"Disparate Impact: {di:.4f} (threshold={DI_THRESHOLD}) {'✓' if di >= DI_THRESHOLD else '✗'}")
    print(f"Gate: {'PASSED' if passed else 'FAILED'}")

    return {"auc": auc, "disparate_impact": di, "passed": passed}
