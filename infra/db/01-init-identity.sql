\connect mesob_identity

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE regions (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(128) NOT NULL,
    code VARCHAR(8)   NOT NULL UNIQUE
);

CREATE TABLE users (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    msisdn            VARCHAR(20)  NOT NULL,
    fan_ciphertext    BYTEA,
    fan_hash          BYTEA,
    full_name_enc     BYTEA,
    kyc_tier          SMALLINT     NOT NULL DEFAULT 0,
    region_id         UUID         REFERENCES regions(id),
    status            VARCHAR(16)  NOT NULL DEFAULT 'PENDING',
    preferred_lang    VARCHAR(8)   NOT NULL DEFAULT 'am',
    wallet_account_id UUID,
    version           INT          NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_users_msisdn ON users(msisdn) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_users_fan    ON users(fan_hash) WHERE deleted_at IS NULL AND fan_hash IS NOT NULL;
CREATE INDEX ix_users_region ON users(region_id);
CREATE INDEX ix_users_status ON users(status);

CREATE TABLE credentials (
    user_id       UUID        PRIMARY KEY REFERENCES users(id),
    pin_hash      BYTEA       NOT NULL,
    failed_count  SMALLINT    NOT NULL DEFAULT 0,
    locked_until  TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE kyc_limits (
    tier            SMALLINT PRIMARY KEY,
    per_txn_minor   BIGINT   NOT NULL,
    daily_minor     BIGINT   NOT NULL,
    balance_minor   BIGINT   NOT NULL
);

INSERT INTO kyc_limits VALUES
    (0, 50000,   200000,   500000),
    (1, 500000,  2000000,  5000000),
    (2, 5000000, 20000000, 50000000);

CREATE TABLE kyc_limits_history (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tier          SMALLINT    NOT NULL,
    per_txn_minor BIGINT      NOT NULL,
    daily_minor   BIGINT      NOT NULL,
    balance_minor BIGINT      NOT NULL,
    valid_from    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to      TIMESTAMPTZ,
    changed_by    UUID,
    change_reason TEXT
);

CREATE TABLE outbox_events (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    topic        VARCHAR(128) NOT NULL,
    aggregate_id VARCHAR(128) NOT NULL DEFAULT '',
    payload      JSONB        NOT NULL,
    published    BOOLEAN      NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_outbox_unpublished ON outbox_events(created_at) WHERE published = false;
