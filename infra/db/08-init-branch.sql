\connect mesob_branch

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE branches (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(128) NOT NULL,
    region_id  UUID         NOT NULL,
    officer_id UUID,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE approvals (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id            UUID        NOT NULL,
    branch_id           UUID        NOT NULL REFERENCES branches(id),
    float_ceiling_minor BIGINT      NOT NULL,
    approved_by         UUID        NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE disputes (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id   UUID        NOT NULL,
    raised_by        UUID        NOT NULL,
    reason           TEXT,
    resolution       VARCHAR(16),
    reversal_txn_id  UUID,
    second_authoriser UUID,
    resolved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
