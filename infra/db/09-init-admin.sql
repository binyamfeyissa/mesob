\connect mesob_admin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE feature_flags (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(64) NOT NULL UNIQUE,
    enabled     BOOLEAN     NOT NULL DEFAULT false,
    rollout_pct SMALLINT    NOT NULL DEFAULT 0,
    updated_by  UUID,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE config_items (
    key        VARCHAR(128) PRIMARY KEY,
    value      JSONB        NOT NULL,
    version    INT          NOT NULL DEFAULT 1,
    updated_by UUID,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE config_history (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key        VARCHAR(128) NOT NULL,
    value      JSONB        NOT NULL,
    version    INT          NOT NULL,
    valid_from TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    valid_to   TIMESTAMPTZ,
    changed_by UUID,
    reason     TEXT
);

CREATE TABLE audit_log (
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    actor_id   UUID,
    actor_role VARCHAR(32),
    action     VARCHAR(64) NOT NULL,
    target     VARCHAR(128),
    channel    VARCHAR(16),
    ip         INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions for audit_log
DO $$
DECLARE
    start_date DATE := DATE_TRUNC('month', NOW());
    end_date   DATE := start_date + INTERVAL '24 months';
    cur_date   DATE := start_date;
    part_name  TEXT;
    next_date  DATE;
BEGIN
    WHILE cur_date < end_date LOOP
        next_date := cur_date + INTERVAL '1 month';
        part_name := 'audit_log_' || TO_CHAR(cur_date, 'YYYY_MM');
        EXECUTE FORMAT(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log FOR VALUES FROM (%L) TO (%L)',
            part_name, cur_date, next_date
        );
        cur_date := next_date;
    END LOOP;
END $$;

CREATE INDEX ix_audit_actor ON audit_log(actor_id, created_at);
CREATE INDEX ix_audit_created ON audit_log USING BRIN (created_at);
