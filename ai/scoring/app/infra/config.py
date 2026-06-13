import os


class Config:
    DB_URL: str = os.getenv("MESOB_SCORING_DB_URL", "postgresql://mesob:mesob@localhost:5432/mesob_scoring")
    REDIS_URL: str = os.getenv("MESOB_SCORING_REDIS_URL", "redis://localhost:6379/0")
    KAFKA_BROKERS: str = os.getenv("MESOB_SCORING_KAFKA_BROKERS", "localhost:9092")
    PORT: int = int(os.getenv("MESOB_SCORING_PORT", "9001"))
    MODEL_PATH: str = os.getenv("MESOB_SCORING_MODEL_PATH", "/models/scoring_model.pkl")


config = Config()
