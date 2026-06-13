\connect mesob_payments

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE merchants (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    name           VARCHAR(128)  NOT NULL,
    status         VARCHAR(16)   NOT NULL DEFAULT 'ACTIVE',
    account_id     UUID,
    commission_pct NUMERIC(5,2)  NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE TABLE billers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(128) NOT NULL,
    adapter_key VARCHAR(64)  NOT NULL UNIQUE,
    status      VARCHAR(16)  NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE payment_refs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    type           VARCHAR(16) NOT NULL,
    transaction_id UUID,
    biller_ref     VARCHAR(128),
    status         VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE outbox_events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id VARCHAR(128) NOT NULL DEFAULT '',
    topic        VARCHAR(128) NOT NULL,
    payload      JSONB        NOT NULL,
    published    BOOLEAN     NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_outbox_unpublished ON outbox_events(created_at) WHERE published = false;
