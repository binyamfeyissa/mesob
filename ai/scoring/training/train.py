"""Train XGBoost credit scoring model.

Features (no protected attributes — no ethnicity, religion, gender, or language):
  iqub_punctuality, savings_velocity, transaction_frequency,
  prior_loan_repaid, prior_loan_defaulted, account_tenure_months
"""
import argparse
import json
import os
import pickle

import numpy as np
import pandas as pd


def load_data(data_path: str) -> pd.DataFrame:
    """Load features + labels from a CSV or JSONL file."""
    if data_path.endswith(".csv"):
        return pd.read_csv(data_path)
    with open(data_path) as f:
        rows = [json.loads(line) for line in f if line.strip()]
    return pd.DataFrame(rows)


FEATURE_COLS = [
    "iqub_punctuality",
    "savings_velocity_minor",
    "transaction_frequency",
    "prior_loan_repaid",
    "prior_loan_defaulted",
    "account_tenure_months",
]

LABEL_COL = "repaid"  # 1=repaid, 0=defaulted


def train(df: pd.DataFrame):
    """Train and return an XGBoost classifier."""
    try:
        import xgboost as xgb
    except ImportError:
        raise RuntimeError("xgboost not installed — run: pip install xgboost")

    available = [c for c in FEATURE_COLS if c in df.columns]
    X = df[available].fillna(0).values
    y = df[LABEL_COL].values

    model = xgb.XGBClassifier(
        n_estimators=300,
        max_depth=4,
        learning_rate=0.05,
        subsample=0.8,
        colsample_bytree=0.8,
        use_label_encoder=False,
        eval_metric="auc",
        random_state=42,
    )
    model.fit(X, y)
    model.feature_names_ = available
    return model


def main():
    parser = argparse.ArgumentParser(description="Train Mesob credit scoring model")
    parser.add_argument("--data-path", required=True, help="Path to training data (CSV or JSONL)")
    parser.add_argument("--output-path", required=True, help="Path to save model pickle")
    parser.add_argument("--min-rows", type=int, default=100, help="Minimum rows required to train")
    args = parser.parse_args()

    df = load_data(args.data_path)
    print(f"Loaded {len(df)} rows from {args.data_path}")

    if len(df) < args.min_rows:
        print(f"Not enough data ({len(df)} < {args.min_rows}) — falling back to rules-only")
        return

    if LABEL_COL not in df.columns:
        print(f"Missing label column '{LABEL_COL}' — aborting")
        return

    model = train(df)

    os.makedirs(os.path.dirname(os.path.abspath(args.output_path)), exist_ok=True)
    with open(args.output_path, "wb") as f:
        pickle.dump(model, f)

    print(f"Model saved → {args.output_path}")
    print(f"Features used: {model.feature_names_}")


if __name__ == "__main__":
    main()
