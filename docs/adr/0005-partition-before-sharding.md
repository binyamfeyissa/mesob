# ADR 0005: Monthly table partitioning before horizontal sharding

**Status**: Accepted

## Context
`ledger_entries`, `transactions`, and `audit_log` will grow very large over time. We needed a strategy for managing table growth without premature complexity.

## Decision
Use PostgreSQL declarative range partitioning by `created_at` month for all high-volume append-only tables. We pre-create 24 monthly partitions at database init. Horizontal sharding (e.g. Citus) is deferred until partition pruning is insufficient.

## Consequences
- **Positive**: Partition pruning makes date-ranged queries dramatically faster.
- **Positive**: Old partitions can be detached and archived without downtime.
- **Positive**: Avoids premature complexity of distributed SQL.
- **Negative**: Queries without a date filter must scan all partitions.
- **Negative**: Application must provide `created_at` in all writes (not generated only at DB level).
