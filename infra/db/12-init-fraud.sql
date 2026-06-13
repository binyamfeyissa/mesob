\connect mesob_fraud

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE fraud_alerts (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL,
    transaction_id UUID,
    rule           VARCHAR(64) NOT NULL,
    severity       VARCHAR(16) NOT NULL DEFAULT 'HIGH',
    risk_score     NUMERIC(5,2) NOT NULL,
    status         VARCHAR(16) NOT NULL DEFAULT 'OPEN',
    disposition    VARCHAR(16),
    note           TEXT,
    disposed_by    UUID,
    disposed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_alerts_status ON fraud_alerts(status) WHERE status = 'OPEN';
CREATE INDEX ix_alerts_user   ON fraud_alerts(user_id, created_at DESC);
