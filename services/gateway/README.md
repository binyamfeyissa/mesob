# Gateway Service

Pure BFF/routing layer. Validates JWT (RS256), rate-limits, proxies to domain services.

Owns: no data store. Reads: JWT public key.

## Local
```bash
MESOB_GATEWAY_IDENTITY_URL=http://localhost:8001 go run ./cmd/server
```
