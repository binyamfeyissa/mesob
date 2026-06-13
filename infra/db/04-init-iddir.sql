\connect mesob_iddir

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE iddir_groups (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(128) NOT NULL,
    premium_minor   BIGINT      NOT NULL,
    frequency       VARCHAR(16) NOT NULL,
    benefit_minor   BIGINT      NOT NULL,
    status          VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    leader_id       UUID        NOT NULL,
    pool_account_id UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE iddir_memberships (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID        NOT NULL REFERENCES iddir_groups(id),
    user_id         UUID        NOT NULL,
    coverage_status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_iddir_membership ON iddir_memberships(group_id, user_id);

CREATE TABLE iddir_premiums (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    membership_id UUID        NOT NULL REFERENCES iddir_memberships(id),
    period        VARCHAR(8)  NOT NULL,
    transaction_id UUID,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE iddir_claims (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id       UUID        NOT NULL REFERENCES iddir_groups(id),
    member_id      UUID        NOT NULL,
    type           VARCHAR(32) NOT NULL,
    description    TEXT,
    evidence_ref   VARCHAR(256),
    status         VARCHAR(16) NOT NULL DEFAULT 'UNDER_REVIEW',
    settled_minor  BIGINT,
    transaction_id UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
