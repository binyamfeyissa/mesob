"""SHAP TreeExplainer wrapper for XGBoost model."""
from typing import Optional
import numpy as np


class SHAPExplainer:
    def __init__(self, model=None):
        self._model = model
        self._explainer = None

    def load(self, model):
        """Load model and initialize SHAP explainer."""
        self._model = model
        try:
            import shap
            self._explainer = shap.TreeExplainer(model)
        except Exception:
            self._explainer = None

    def explain(self, feature_vector: dict) -> list[dict]:
        """Return SHAP factors for a feature vector. Returns [] if explainer unavailable."""
        if self._explainer is None:
            return []
        try:
            import shap
            import numpy as np
            X = np.array([[v for v in feature_vector.values()]])
            shap_values = self._explainer.shap_values(X)
            factors = []
            for i, (k, v) in enumerate(feature_vector.items()):
                factors.append({"feature": k, "contribution": float(shap_values[0][i])})
            return sorted(factors, key=lambda x: abs(x["contribution"]), reverse=True)
        except Exception:
            return []
