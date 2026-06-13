# Identity Service

Owns: user registration, PIN auth, JWT session issuance, KYC tier management.
Owns tables: users, credentials, kyc_limits (in mesob_identity DB).
Emits: UserActivated, KycTierChanged.

## Local
```bash
MESOB_IDENTITY_DB_URL="postgres://mesob:mesob@localhost:5432/mesob_identity?sslmode=disable" \
go run ./cmd/server
```
