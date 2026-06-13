# ADR 0006: Stateless JWT access token + opaque refresh token

**Status**: Accepted

## Context
Authentication must be stateless at the Gateway (no DB hit per request) to meet the USSD ≤2s response budget. But we also need the ability to revoke sessions.

## Decision
- **Access token**: Short-lived (15 min), RS256/EdDSA JWT verified at the Gateway using the public key only. No DB hit.
- **Refresh token**: 7-day opaque random token stored in Redis (hashed). Exchanged for a new access token at `/identity/token/refresh`.
- **JWT payload**: `sub`, `role`, `scope`, `kyc_tier`, `region_id`, `exp`, `jti`.

## Consequences
- **Positive**: Gateway validates tokens in-process — zero DB round-trips. USSD budget met.
- **Positive**: Refresh tokens can be revoked immediately by deleting from Redis.
- **Negative**: Access tokens cannot be revoked mid-lifetime (15min window). Acceptable given the short TTL.
