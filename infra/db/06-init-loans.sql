\connect mesob_loans

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE credit_scores (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL,
    score          SMALLINT    NOT NULL,
    tier           CHAR(1)     NOT NULL,
    ceiling_minor  BIGINT      NOT NULL,
    model_ver      VARCHAR(64) NOT NULL,
    source         VARCHAR(16) NOT NULL DEFAULT 'ML',
    shap_json      JSONB,
    inputs_hash    BYTEA,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_scores_user_latest ON credit_scores(user_id, created_at DESC);

CREATE TABLE loans (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL,
    principal_minor  BIGINT      NOT NULL,
    fee_minor        BIGINT      NOT NULL DEFAULT 0,
    outstanding_minor BIGINT     NOT NULL,
    score_id         UUID        REFERENCES credit_scores(id),
    status           VARCHAR(24) NOT NULL DEFAULT 'ACTIVE',
    mode             VARCHAR(16) NOT NULL DEFAULT 'INTERNAL',
    due_date         DATE        NOT NULL,
    mfi_facility_id  VARCHAR(128),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_loans_user   ON loans(user_id);
CREATE INDEX ix_loans_status ON loans(status);
CREATE INDEX ix_loans_due    ON loans(status, due_date) WHERE status = 'ACTIVE';

CREATE TABLE repayments (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id        UUID        NOT NULL REFERENCES loans(id),
    amount_minor   BIGINT      NOT NULL,
    transaction_id UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE decisions (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id        UUID        REFERENCES loans(id),
    score_id       UUID        REFERENCES credit_scores(id),
    decision       VARCHAR(16) NOT NULL,
    reasons        JSONB,
    ceiling_minor  BIGINT,
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
