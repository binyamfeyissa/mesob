from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from app.api.routes import router
from app.infra.config import Config


@asynccontextmanager
async def lifespan(app: FastAPI):
    # startup: load anomaly model, connect redis
    yield
    # shutdown


def create_app() -> FastAPI:
    app = FastAPI(
        title="Mesob Fraud/AML Screening",
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
    cfg = Config()
    uvicorn.run("app.main:app", host="0.0.0.0", port=int(cfg.HTTP_PORT), reload=False)
