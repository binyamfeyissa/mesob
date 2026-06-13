# ADR 0002: gRPC internal, REST at edge

**Status**: Accepted

## Context
We have two communication patterns: service-to-service (internal) and client-to-service (external). We needed a protocol choice for each.

## Decision
- Internal service-to-service: gRPC with mTLS. Typed contracts via `.proto` files in `shared/proto/`.
- External (client-facing): REST/JSON via the API Gateway. Clients receive structured error envelopes.

## Consequences
- **Positive**: Strong typing and efficient binary encoding for internal calls.
- **Positive**: REST is universally supported by mobile and web clients.
- **Positive**: gRPC mTLS enforces that only authorised services can call internal endpoints.
- **Negative**: Two protocols to maintain and document.
