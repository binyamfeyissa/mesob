# Mesob Wallet — Monorepo

Ethiopian last-mile fintech platform. Offline-first mobile money, rotating savings (Iqub), mutual insurance (Iddir), microlending, agent network.

## Quick Start

Prerequisites: Go 1.25+, Node 22+, Python 3.11+, Docker

```bash
make dev          # start Postgres, Redis, Kafka
make build        # compile all Go services
make test         # run domain-layer unit tests
make proto        # regenerate gRPC stubs
```

## Services

| Service | Port | gRPC |
|---------|------|------|
| gateway | 8000 | — |
| identity | 8001 | 9101 |
| ledger | 8002 | 9102 |
| iqub | 8003 | 9103 |
| iddir | 8004 | 9104 |
| agent | 8005 | 9105 |
| loans | 8006 | 9106 |
| payments | 8007 | 9107 |
| branch | 8008 | 9108 |
| admin | 8009 | 9109 |
| ussd | 8010 | 9110 |
| telegram | 8011 | 9111 |
| notification | 8012 | 9112 |
| adapter-hub | 8013 | 9113 |
| scoring (Python) | 9001 | — |
| fraud (Python) | 9002 | — |
| web-console | 3000 | — |
