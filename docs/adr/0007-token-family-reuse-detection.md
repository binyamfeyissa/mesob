# ADR 0007: Token family rotation and reuse detection

**Status**: Accepted

## Context
If a refresh token is stolen, an attacker can use it to obtain new access tokens. We needed a mechanism to detect stolen refresh tokens.

## Decision
Implement **token family rotation**: each refresh token belongs to a family. When a refresh is performed, the old token is invalidated and a new token is issued (same family). If a previously-used (already-invalidated) token is presented, the **entire family is invalidated** immediately — both the attacker's session and the legitimate user's session. The user is forced to re-authenticate.

## Consequences
- **Positive**: Stolen token reuse is detected and both sessions are immediately terminated.
- **Positive**: Legitimate users get a clear signal (session expired) to re-authenticate and change their PIN.
- **Negative**: A network error during refresh can cause the client to retry with an already-rotated token, triggering false-positive family invalidation. Mitigated by retry logic with idempotency.
