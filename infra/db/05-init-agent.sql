\connect mesob_agent

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE agents (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL UNIQUE,
    float_account_id  UUID,
    float_limit_minor BIGINT      NOT NULL DEFAULT 100000000,
    region_id         UUID        NOT NULL,
    status            VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX ix_agents_region ON agents(region_id);
CREATE INDEX ix_agents_status ON agents(status);

CREATE TABLE settlements (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id         UUID        NOT NULL REFERENCES agents(id),
    branch_id        UUID,
    amount_minor     BIGINT      NOT NULL,
    transaction_id   UUID,
    authorised_by    UUID        NOT NULL,
    second_authoriser UUID,
    confirmed_at     TIMESTAMPTZ,
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
