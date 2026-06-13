-- Demo account seed for mesob_identity database.
-- Static (non-PIN) data only. PIN hashes are inserted by `infra/scripts/seed_demo`.
-- Run AFTER all migration scripts (00..13).
--
-- Demo accounts:
--   SUPER_ADMIN : +251911000001 / PIN 111111
--   ADMIN       : +251911000002 / PIN 111111
--   BRANCH_MGR  : +251911000003 / PIN 111111
--   AGENT       : +251911000004 / PIN 111111
--   USER (T2)   : +251911000005 / PIN 111111
--   USER (T0)   : +251911000006 / PIN 111111

\connect mesob_identity

-- Seed regions (idempotent)
INSERT INTO regions (id, name, code) VALUES
    ('00000000-0000-7000-8000-000000000001', 'Addis Ababa',  'AA'),
    ('00000000-0000-7000-8000-000000000002', 'Oromia',       'OR'),
    ('00000000-0000-7000-8000-000000000003', 'Amhara',       'AM'),
    ('00000000-0000-7000-8000-000000000004', 'Tigray',       'TI'),
    ('00000000-0000-7000-8000-000000000005', 'SNNPR',        'SN')
ON CONFLICT (code) DO NOTHING;

-- Add role column if not present (demo convenience — production derives role from staff tables)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'USER';

-- Demo users (PIN hashes are empty here — the Go seeder fills them)
INSERT INTO users (id, msisdn, kyc_tier, region_id, status, preferred_lang, role) VALUES
    ('01000000-0000-7000-8000-000000000001', '+251911000001', 2, '00000000-0000-7000-8000-000000000001', 'ACTIVE', 'en', 'SUPER_ADMIN'),
    ('01000000-0000-7000-8000-000000000002', '+251911000002', 2, '00000000-0000-7000-8000-000000000001', 'ACTIVE', 'en', 'ADMIN'),
    ('01000000-0000-7000-8000-000000000003', '+251911000003', 1, '00000000-0000-7000-8000-000000000001', 'ACTIVE', 'am', 'BRANCH_MANAGER'),
    ('01000000-0000-7000-8000-000000000004', '+251911000004', 1, '00000000-0000-7000-8000-000000000002', 'ACTIVE', 'am', 'AGENT'),
    ('01000000-0000-7000-8000-000000000005', '+251911000005', 2, '00000000-0000-7000-8000-000000000001', 'ACTIVE', 'am', 'USER'),
    ('01000000-0000-7000-8000-000000000006', '+251911000006', 0, '00000000-0000-7000-8000-000000000003', 'ACTIVE', 'am', 'USER')
ON CONFLICT DO NOTHING;

-- Add wallet_account_id column for existing dev databases (idempotent)
ALTER TABLE users ADD COLUMN IF NOT EXISTS wallet_account_id UUID;

-- Placeholder credentials — will be overwritten by `make seed-demo` with real Argon2id hashes
-- The Go seeder inserts/upserts these with properly hashed PINs.

-- ─────────────────────────────────────────────────────────────────────────────
-- Ledger accounts
-- ─────────────────────────────────────────────────────────────────────────────
\connect mesob_ledger

-- Float account for demo AGENT (user_id 01000000-0000-7000-8000-000000000004)
INSERT INTO accounts (id, owner_type, owner_id, type, currency, status, allow_negative) VALUES
    ('00000000-0000-7000-8000-100000000001',
     'AGENT', '01000000-0000-7000-8000-000000000004',
     'FLOAT', 'ETB', 'ACTIVE', false)
ON CONFLICT (id) DO NOTHING;

-- Wallet accounts for demo users
INSERT INTO accounts (id, owner_type, owner_id, type, currency, status, allow_negative) VALUES
    ('00000000-0000-7000-8000-200000000005',
     'USER', '01000000-0000-7000-8000-000000000005',
     'WALLET', 'ETB', 'ACTIVE', false),
    ('00000000-0000-7000-8000-200000000006',
     'USER', '01000000-0000-7000-8000-000000000006',
     'WALLET', 'ETB', 'ACTIVE', false)
ON CONFLICT (id) DO NOTHING;

-- Initial balances
INSERT INTO account_balances (account_id, balance_minor) VALUES
    ('00000000-0000-7000-8000-100000000001', 50000000),   -- agent float: 500 ETB
    ('00000000-0000-7000-8000-200000000005', 10000000),   -- user 005 wallet: 100 ETB
    ('00000000-0000-7000-8000-200000000006', 5000000)     -- user 006 wallet: 50 ETB
ON CONFLICT (account_id) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Agent record
-- ─────────────────────────────────────────────────────────────────────────────
\connect mesob_agent

INSERT INTO agents (id, user_id, float_account_id, float_limit_minor, region_id, status) VALUES
    ('00000000-0000-7000-8000-300000000001',
     '01000000-0000-7000-8000-000000000004',
     '00000000-0000-7000-8000-100000000001',
     100000000,
     '00000000-0000-7000-8000-000000000002',
     'ACTIVE')
ON CONFLICT (user_id) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Link wallet accounts back to identity users
-- ─────────────────────────────────────────────────────────────────────────────
\connect mesob_identity

UPDATE users SET wallet_account_id = '00000000-0000-7000-8000-200000000005'
    WHERE id = '01000000-0000-7000-8000-000000000005';
UPDATE users SET wallet_account_id = '00000000-0000-7000-8000-200000000006'
    WHERE id = '01000000-0000-7000-8000-000000000006';
