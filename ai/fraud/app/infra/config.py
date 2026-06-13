import os


class Config:
    HTTP_PORT: str = os.getenv("MESOB_FRAUD_HTTP_PORT", "9002")
    DB_URL: str = os.getenv(
        "MESOB_FRAUD_DB_URL",
        "postgresql://mesob:mesob@localhost:5432/mesob_fraud",
    )
    REDIS_URL: str = os.getenv("MESOB_FRAUD_REDIS_URL", "redis://localhost:6379/1")
    KAFKA_BROKERS: str = os.getenv("MESOB_FRAUD_KAFKA_BROKERS", "localhost:9092")
