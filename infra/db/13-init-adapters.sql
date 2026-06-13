\connect mesob_adapters

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE adapter_requests (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    adapter        VARCHAR(32) NOT NULL,
    idempotency_key VARCHAR(64),
    request_json   JSONB,
    response_json  JSONB,
    status         VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_adapter_idem ON adapter_requests(adapter, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
