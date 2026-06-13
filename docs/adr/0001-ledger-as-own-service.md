# ADR 0001: Ledger as its own microservice

**Status**: Accepted

## Context
Every money movement in Mesob Wallet requires a double-entry accounting record. We needed to decide whether ledger logic lives inside each domain service or as a shared capability.

## Decision
The Ledger is a standalone service (`services/ledger`) with its own PostgreSQL database. All money movements are expressed as balanced transaction entries via the Ledger's internal API. No other service may write directly to ledger tables.

## Consequences
- **Positive**: Single source of truth for all balances and movements. Easy audit, reconciliation, and reversal.
- **Positive**: Balance sheet is always globally consistent — no cross-service sync needed.
- **Negative**: Additional network hop for every money operation.
- **Negative**: Ledger becomes a dependency for Payments, Iqub, Iddir, Loans, Agent — it must be highly available.
