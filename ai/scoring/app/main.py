import asyncio
import json
import os
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from app.api.routes import router
from app.consumers.iqub import handle_iqub_contribution_missed, handle_iqub_contribution_recorded
from app.consumers.labels import handle_loan_defaulted, handle_loan_repaid
from app.consumers.transaction import handle_transaction_posted
from app.infra.config import config

TOPIC_HANDLERS = {
    "ledger.transaction-posted":    handle_transaction_posted,
    "iqub.contribution-recorded":   handle_iqub_contribution_recorded,
    "iqub.contribution-missed":     handle_iqub_contribution_missed,
    "loans.repaid":                 handle_loan_repaid,
    "loans.defaulted":              handle_loan_defaulted,
}


async def _run_consumer(brokers: str, topics: list[str]) -> None:
    """Background asyncio task consuming domain events into the feature store."""
    try:
        from kafka import KafkaConsumer  # type: ignore
        consumer = KafkaConsumer(
            *topics,
            bootstrap_servers=brokers.split(","),
            group_id="scoring-feature-updater",
            auto_offset_reset="latest",
            value_deserializer=lambda b: json.loads(b.decode("utf-8")),
            consumer_timeout_ms=1000,
        )
    except Exception:
        return  # Kafka unavailable — feature store updates via HTTP fallback only

    loop = asyncio.get_event_loop()
    while True:
        try:
            for message in consumer:
                handler = TOPIC_HANDLERS.get(message.topic)
                if handler:
                    try:
                        await loop.run_in_executor(None, handler, message.value)
                    except Exception:
                        pass
        except Exception:
            pass
        await asyncio.sleep(0.1)


@asynccontextmanager
async def lifespan(app: FastAPI):
    brokers = config.KAFKA_BROKERS
    topics = list(TOPIC_HANDLERS.keys())
    task = asyncio.create_task(_run_consumer(brokers, topics))
    yield
    task.cancel()
    try:
        await task
    except asyncio.CancelledError:
        pass


def create_app() -> FastAPI:
    app = FastAPI(
        title="Mesob Credit Scoring",
        version="0.1.0",
        lifespan=lifespan,
    )
    app.include_router(router)

    @app.get("/health")
    def health():
        return {"status": "ok"}

    @app.get("/ready")
    def ready():
        return {"status": "ready"}

    return app


app = create_app()

if __name__ == "__main__":
    uvicorn.run("app.main:app", host="0.0.0.0", port=config.PORT, reload=False)
