# ADR 0003: Transactional outbox pattern for Kafka events

**Status**: Accepted

## Context
Distributed systems face the dual-write problem: writing to a database and publishing to Kafka in the same operation cannot be made atomic with a standard 2PC.

## Decision
Every service that emits events writes to an `outbox_events` table in the same database transaction as its business write. A relay worker polls the table and publishes to Kafka, marking rows as published after a successful broker acknowledgement.

## Consequences
- **Positive**: No lost events — business state and event emission are atomic.
- **Positive**: Exactly-once delivery semantics achievable with idempotent consumers (dedupe on `event_id`).
- **Negative**: Increased DB write amplification (one extra row per event).
- **Negative**: Relay worker adds operational complexity and must be monitored.
