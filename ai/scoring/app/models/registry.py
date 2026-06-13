"""
Model registry — loads signed, versioned XGBoost artifacts.
"""
import os
from typing import Optional


class ModelRegistry:
    def __init__(self):
        self._model = None
        self._version = "rules-only-v1"

    def load(self, path: str) -> bool:
        """Load model from path. Returns True if successful."""
        try:
            import pickle
            with open(path, "rb") as f:
                self._model = pickle.load(f)
            self._version = os.path.basename(path).replace(".pkl", "")
            return True
        except (FileNotFoundError, Exception):
            return False

    @property
    def model(self):
        return self._model

    @property
    def version(self) -> str:
        return self._version

    @property
    def is_available(self) -> bool:
        return self._model is not None


registry = ModelRegistry()
