# Ledger Service

Owns: double-entry ledger. Immutable, append-only entries. Internal service only (mTLS).
Tables: accounts, transactions (partitioned), ledger_entries (partitioned), account_balances, outbox_events, idempotency_keys.
Emits: TransactionPosted, TransactionReversed.

## Invariants
- SUM(debits) == SUM(credits) enforced by assert_balanced() DB trigger
- All amount_minor > 0 (no float64)
- Idempotency-Key required on all POST /transactions calls

## Local
```bash
MESOB_LEDGER_DB_URL="postgres://mesob:mesob@localhost:5432/mesob_ledger?sslmode=disable" \
go run ./cmd/server
```
