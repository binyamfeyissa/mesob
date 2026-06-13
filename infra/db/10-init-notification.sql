\connect mesob_notification

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE templates (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key        VARCHAR(64) NOT NULL,
    lang       VARCHAR(8)  NOT NULL,
    channel    VARCHAR(16) NOT NULL,
    body       TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(key, lang, channel)
);

INSERT INTO templates (key, lang, channel, body) VALUES
    ('otp', 'am', 'SMS', 'የ Mesob Wallet ማረጋገጫ ኮድዎ: {{code}}'),
    ('otp', 'en', 'SMS', 'Your Mesob Wallet verification code: {{code}}'),
    ('otp', 'om', 'SMS', 'Koodii mirkaneessa Mesob Wallet keessan: {{code}}'),
    ('otp', 'ti', 'SMS', 'ናይ Mesob Wallet መረጋገጺ ኮድካ: {{code}}'),
    ('txn_credit', 'am', 'SMS', 'ብር {{amount}} ወደ ቆጠባዎ ገብቷል። ቀሪ ቀሪ: {{balance}}'),
    ('txn_credit', 'en', 'SMS', 'ETB {{amount}} credited. New balance: {{balance}}'),
    ('txn_debit', 'am', 'SMS', 'ብር {{amount}} ከቆጠባዎ ወጥቷል። ቀሪ ቀሪ: {{balance}}'),
    ('txn_debit', 'en', 'SMS', 'ETB {{amount}} debited. New balance: {{balance}}');

CREATE TABLE delivery_log (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL,
    template_key  VARCHAR(64) NOT NULL,
    channel       VARCHAR(16) NOT NULL,
    status        VARCHAR(16) NOT NULL DEFAULT 'QUEUED',
    attempts      SMALLINT    NOT NULL DEFAULT 0,
    last_error    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_delivery_user ON delivery_log(user_id, created_at DESC);
CREATE INDEX ix_delivery_status ON delivery_log(status) WHERE status != 'DELIVERED';
