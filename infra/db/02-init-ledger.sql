\connect mesob_ledger

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE accounts (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type     VARCHAR(16) NOT NULL,
    owner_id       UUID        NOT NULL,
    type           VARCHAR(20) NOT NULL,
    currency       CHAR(3)     NOT NULL DEFAULT 'ETB',
    status         VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    allow_negative BOOLEAN     NOT NULL DEFAULT false,
    version        INT         NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_accounts_owner ON accounts(owner_type, owner_id);

CREATE TABLE transactions (
    id               UUID        NOT NULL DEFAULT gen_random_uuid(),
    type             VARCHAR(24) NOT NULL,
    status           VARCHAR(16) NOT NULL DEFAULT 'POSTED',
    idempotency_key  VARCHAR(64) NOT NULL,
    initiated_by     UUID,
    channel          VARCHAR(16),
    reverses_txn_id  UUID,           -- app-level ref; no FK (cross-partition not supported)
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)    -- partition key must be in PK
) PARTITION BY RANGE (created_at);

-- Unique per idempotency_key within each partition (month-scoped; global dedup via idempotency_keys table)
CREATE UNIQUE INDEX uq_txn_idem ON transactions(idempotency_key, created_at);
CREATE INDEX ix_txn_created ON transactions USING BRIN (created_at);

-- Create monthly partitions for the next 2 years
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
        part_name := 'transactions_' || TO_CHAR(cur_date, 'YYYY_MM');
        EXECUTE FORMAT(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF transactions FOR VALUES FROM (%L) TO (%L)',
            part_name, cur_date, next_date
        );
        cur_date := next_date;
    END LOOP;
END $$;

CREATE TABLE ledger_entries (
    id             UUID        NOT NULL DEFAULT gen_random_uuid(),
    transaction_id UUID        NOT NULL,
    account_id     UUID        NOT NULL REFERENCES accounts(id),
    direction      CHAR(1)     NOT NULL CHECK (direction IN ('D', 'C')),
    amount_minor   BIGINT      NOT NULL CHECK (amount_minor > 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX ix_entries_txn     ON ledger_entries(transaction_id, created_at);
CREATE INDEX ix_entries_account ON ledger_entries(account_id, created_at DESC);

-- Create monthly partitions for ledger_entries
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
        part_name := 'ledger_entries_' || TO_CHAR(cur_date, 'YYYY_MM');
        EXECUTE FORMAT(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF ledger_entries FOR VALUES FROM (%L) TO (%L)',
            part_name, cur_date, next_date
        );
        cur_date := next_date;
    END LOOP;
END $$;

CREATE TABLE account_balances (
    account_id   UUID        PRIMARY KEY REFERENCES accounts(id),
    balance_minor BIGINT     NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
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

CREATE TABLE idempotency_keys (
    key            VARCHAR(64) PRIMARY KEY,
    transaction_id UUID        NOT NULL,    -- no FK; transactions is partitioned
    response_hash  BYTEA,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at     TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

-- Trigger to enforce double-entry balance
CREATE OR REPLACE FUNCTION assert_balanced()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    txn_id UUID;
    debit_sum BIGINT;
    credit_sum BIGINT;
BEGIN
    txn_id := NEW.transaction_id;
    SELECT
        COALESCE(SUM(CASE WHEN direction = 'D' THEN amount_minor ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN direction = 'C' THEN amount_minor ELSE 0 END), 0)
    INTO debit_sum, credit_sum
    FROM ledger_entries
    WHERE transaction_id = txn_id;

    IF debit_sum != credit_sum THEN
        RAISE EXCEPTION 'UNBALANCED: debits=% credits=% for transaction %', debit_sum, credit_sum, txn_id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE CONSTRAINT TRIGGER trg_assert_balanced
AFTER INSERT ON ledger_entries
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION assert_balanced();
