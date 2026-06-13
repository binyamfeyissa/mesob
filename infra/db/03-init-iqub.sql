\connect mesob_iqub

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE iqub_groups (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(128) NOT NULL,
    cycle_minor     BIGINT      NOT NULL,
    frequency       VARCHAR(16) NOT NULL,
    member_limit    SMALLINT    NOT NULL DEFAULT 20,
    payout_order    VARCHAR(16) NOT NULL DEFAULT 'RANDOM',
    status          VARCHAR(16) NOT NULL DEFAULT 'FORMING',
    leader_id       UUID        NOT NULL,
    agent_id        UUID,
    pool_account_id UUID,
    join_code       VARCHAR(16) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE iqub_memberships (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     UUID        NOT NULL REFERENCES iqub_groups(id),
    user_id      UUID        NOT NULL,
    payout_order SMALLINT,
    cycle_state  VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_memberships_group ON iqub_memberships(group_id);
CREATE INDEX ix_memberships_user  ON iqub_memberships(user_id);
CREATE UNIQUE INDEX uq_membership ON iqub_memberships(group_id, user_id);

CREATE TABLE iqub_cycles (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     UUID        NOT NULL REFERENCES iqub_groups(id),
    number       SMALLINT    NOT NULL,
    status       VARCHAR(16) NOT NULL DEFAULT 'OPEN',
    recipient_id UUID,
    due_date     DATE        NOT NULL,
    closed_at    TIMESTAMPTZ
);

CREATE TABLE outbox_events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    topic        VARCHAR(128) NOT NULL,
    aggregate_id VARCHAR(128) NOT NULL DEFAULT '',
    payload      JSONB       NOT NULL,
    published    BOOLEAN     NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_outbox_unpublished ON outbox_events(created_at) WHERE published = false;
