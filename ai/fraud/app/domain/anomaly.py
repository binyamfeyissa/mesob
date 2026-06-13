"""Isolation Forest anomaly detector for novel fraud patterns."""
from typing import Optional

try:
    import numpy as np
    _NUMPY_AVAILABLE = True
except ImportError:
    _NUMPY_AVAILABLE = False


class AnomalyDetector:
    def __init__(self):
        self._model = None

    def load(self, model_path: str) -> None:
        """Load a trained Isolation Forest model."""
        try:
            import pickle
            with open(model_path, "rb") as f:
                self._model = pickle.load(f)
        except Exception:
            self._model = None

    def score(self, features: dict) -> Optional[float]:
        """Return anomaly score [0,1]. Higher = more anomalous. None if model not loaded."""
        if self._model is None:
            return None
        try:
            X = np.array([[features.get(k, 0.0) for k in sorted(features.keys())]])
            # IsolationForest.score_samples returns negative; convert to [0,1]
            raw = self._model.score_samples(X)[0]
            normalized = 1.0 - (raw + 0.5)  # rough normalization
            return float(max(0.0, min(1.0, normalized)))
        except Exception:
            return None


# Singleton detector loaded at startup
detector = AnomalyDetector()
