\connect mesob_scoring

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE credit_scores (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL,
    score         SMALLINT    NOT NULL,
    tier          CHAR(1)     NOT NULL,
    ceiling_minor BIGINT      NOT NULL,
    model_ver     VARCHAR(64) NOT NULL,
    source        VARCHAR(16) NOT NULL DEFAULT 'ML',
    shap_json     JSONB,
    inputs_hash   BYTEA,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_scores_user_latest ON credit_scores(user_id, created_at DESC);
