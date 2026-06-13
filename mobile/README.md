# mobile

React Native mobile apps for Mesob Wallet.

## Apps

| App | Description | Audience |
|-----|-------------|----------|
| `apps/user` | Wallet, Iqub, Iddir, Loans | Smartphone users |
| `apps/agent` | Cash-in/out, offline sync, float | Field agents |

## Offline sync (Agent app)
- UUIDv7 idempotency keys generated at **capture time** (not sync time)
- Operations queued locally; survive app restarts
- `/agent/sync` batch endpoint used when connectivity available
- Server-authoritative: float ceiling and limit breaches rejected per-op

## Languages
am (Amharic), om (Oromo), ti (Tigrinya), en (English)

## Local setup
```bash
cd mobile
yarn install
yarn user   # start user app
yarn agent  # start agent app
```
