# ADR 0004: Server-authoritative offline sync for agents

**Status**: Accepted

## Context
Field agents in low-connectivity areas need to capture cash-in and cash-out operations offline. We needed a conflict resolution strategy for when operations are synced.

## Decision
Money operations are **never auto-merged**. The agent app captures a queue of intents with UUIDv7 keys generated at capture time (not sync time). At sync, the server validates each operation individually against current state (float ceiling, KYC limits, fraud rules). The server's decision is final.

## Consequences
- **Positive**: No double-spend, no phantom credit. Ledger balance is always correct.
- **Positive**: Idempotency keys generated at capture time make replay safe.
- **Negative**: Operations may be rejected at sync time — the agent must communicate rejections to customers.
- **Negative**: Agent UX must clearly communicate "pending" vs "confirmed" states.
