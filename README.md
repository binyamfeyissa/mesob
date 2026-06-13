# Mesob Wallet Monorepo

Mesob Wallet is an Ethiopian last-mile fintech platform built as a polyglot monorepo.

Core product areas:

- Offline-first mobile money
- Rotating savings circles (Iqub/Equb)
- Mutual insurance (Iddir)
- Agent network and branch operations
- Micro-loans with AI scoring and fraud screening
- Multi-channel access (mobile, web console, USSD, Telegram)

This document is the operational map for the whole repository: architecture, folder ownership, services, local development, testing, and day-to-day commands.

## 1. Technology Stack

- Backend: Go microservices (HTTP + gRPC)
- AI services: Python/FastAPI (`scoring`, `fraud`)
- Web admin console: Next.js 14 (App Router)
- Mobile apps: React Native/Expo workspaces (`user`, `agent`)
- Infra (dev): Docker Compose (Postgres, Redis, Kafka + all services)
- Contracts: OpenAPI + Protobuf

## 2. Repository Layout

```text
mesob/
	ai/                 # Python AI services (scoring, fraud)
	docs/               # Combined OpenAPI spec, runbooks, ADRs
	infra/              # Database bootstrap SQL and helper scripts
	mobile/             # React Native monorepo (user app, agent app)
	services/           # Go microservices (14 services)
	shared/             # Shared Go modules, events, proto, OpenAPI, TS types
	web/                # Next.js web console
	docker-compose.yml  # Full local stack
	Dockerfile.go.dev   # Go dev container image (air live reload)
	go.work             # Go workspace (includes seed CLI)
	go.work.dev         # Go workspace for compose services
	Makefile            # Primary local automation entrypoint
	turbo.json          # Turbo task graph for JS/TS workspaces
```

## 3. What Each Top-Level Folder Owns

### `services/`

All Go business-domain and channel services.

Each service typically contains:

- `cmd/` entrypoints
- `internal/` application/domain logic
- `api/` OpenAPI definitions
- `README.md` service-level notes

### `ai/`

Python services used by lending and payments flows:

- `scoring/` credit scoring and fallback chain
- `fraud/` real-time fraud/AML screening
- `vendor/` offline wheelhouse used by AI builds

### `shared/`

Reusable contracts and utilities across services:

- `go-kit/` common Go middleware/utilities (`auth`, `logging`, `money`, etc.)
- `events/` event envelope + catalog types
- `proto/` shared protobuf contracts (`identity`, `ledger`, `scoring`, etc.)
- `openapi/` shared API docs assets
- `ts-types/` shared TypeScript types package

### `infra/`

Local environment bootstrap and support scripts:

- `db/` ordered SQL setup scripts (`00`..`13`) for per-service databases
- `db/seeds/` demo/test seed SQL
- `scripts/gen-proto.sh` protobuf generation helper
- `scripts/seed_demo` Go seeder (e.g., hashed PIN demo users)

### `docs/`

- `openapi-combined.yaml` full combined API contract
- `swagger-initializer.js` custom Swagger UI setup
- `adr/` architecture decision records
- `runbooks/` operational runbooks

### `mobile/`

React Native workspace:

- `apps/user` smartphone customer app
- `apps/agent` field agent app (offline queue/sync)

### `web/`

- `console/` Next.js 14 admin + branch operations console

## 4. Runtime Architecture at a Glance

```text
Clients
	- Mobile user app
	- Mobile agent app
	- Web console
	- USSD/Telegram channels
				 |
				 v
			gateway (BFF/routing, JWT validation, rate limiting)
				 |
				 v
Domain services (identity, ledger, iqub, iddir, loans, payments, agent, ...)
				 |
				 +--> scoring (AI)
				 +--> fraud (AI, fail-closed callers)
				 +--> adapter-hub (partner integrations)

Infrastructure backbone
	- Postgres (service-owned DBs)
	- Redis (sessions/cache/coordination)
	- Kafka (event stream)
```

Design principles reflected in code and docs:

- Service-owned data boundaries
- Event-driven integration
- Idempotency on critical money flows
- No float arithmetic for money (`amount_minor` style integer units)
- Fail-closed behavior for risk-sensitive integrations

## 5. Service Catalog

### Core financial and identity services

| Service    | HTTP | gRPC | Ownership summary                                              |
| ---------- | ---: | ---: | -------------------------------------------------------------- |
| `gateway`  | 8000 |    - | BFF/routing layer; JWT validation, rate limiting, proxy fanout |
| `identity` | 8001 | 9101 | Registration, PIN auth, JWT session issuance, KYC tier limits  |
| `ledger`   | 8002 | 9102 | Double-entry immutable ledger, balances, posting invariants    |
| `iqub`     | 8003 | 9103 | Rotating savings groups, memberships, cycles                   |
| `iddir`    | 8004 | 9104 | Mutual insurance groups, premiums, claims                      |
| `agent`    | 8005 | 9105 | Agent network cash-in/cash-out and settlements                 |
| `loans`    | 8006 | 9106 | Loan decisioning/disbursement/repayments; uses scoring + fraud |
| `payments` | 8007 | 9107 | P2P, merchant and bill payments; fraud-screened                |
| `branch`   | 8008 | 9108 | Branch workflows (approvals, settlement review, disputes)      |
| `admin`    | 8009 | 9109 | Platform admin, feature flags, config history, audit log       |

### Channel and integration services

| Service        | HTTP | gRPC | Ownership summary                                                |
| -------------- | ---: | ---: | ---------------------------------------------------------------- |
| `ussd`         | 8010 | 9110 | USSD FSM session orchestration (Redis-backed short TTL sessions) |
| `telegram`     | 8011 | 9111 | Telegram webhook bot adapter to gateway/domain flows             |
| `notification` | 8012 | 9112 | Multi-channel notifications (OTP + event-driven dispatch)        |
| `adapter-hub`  | 8013 | 9113 | External integration adapters (NID/MFI) with demo/live modes     |

### AI services

| Service   | HTTP | gRPC | Ownership summary                                          |
| --------- | ---: | ---: | ---------------------------------------------------------- |
| `scoring` | 9001 |    - | Credit scoring with fallback chain (model -> rules -> 422) |
| `fraud`   | 9002 |    - | Real-time fraud/AML screening (ALLOW/REVIEW/BLOCK)         |

### Frontend

| Component     | Port | Purpose                                     |
| ------------- | ---: | ------------------------------------------- |
| `web-console` | 3000 | Admin and branch operator UI                |
| `swagger-ui`  | 8080 | Unified API docs from combined OpenAPI spec |

## 6. Local Development Modes

### A) Fast infra-only boot (recommended first step)

```bash
make dev
```

Starts local infra dependencies used by services:

- Postgres (`localhost:5433`)
- Redis (`localhost:6379`)
- Kafka (`localhost:9092`)

### B) Full stack in containers (all services + apps)

```bash
make all-up
```

Equivalent to:

```bash
docker compose up -d
```

### C) Build and test Go services from host

```bash
make build
make test
make lint
```

What these do:

- `build`: loops all Go services and runs `go build ./...`
- `test`: runs `go test` on `internal/domain/...` and `internal/app/...`
- `lint`: runs `go vet ./...` per service

### D) API contract workflows

```bash
make proto
make swagger-ui
```

- `make proto` runs `infra/scripts/gen-proto.sh`
- `make swagger-ui` serves docs at `http://localhost:8080` using `shared/openapi`

## 7. Databases and Seeding

The compose bootstrap mounts `infra/db` into Postgres init scripts.

Database initialization order in `infra/db`:

- `00-create-databases.sh`
- `01-init-identity.sql`
- `02-init-ledger.sql`
- `03-init-iqub.sql`
- `04-init-iddir.sql`
- `05-init-agent.sql`
- `06-init-loans.sql`
- `07-init-payments.sql`
- `08-init-branch.sql`
- `09-init-admin.sql`
- `10-init-notification.sql`
- `11-init-scoring.sql`
- `12-init-fraud.sql`
- `13-init-adapters.sql`

Demo account seeding:

```bash
make seed-demo
```

This runs SQL seed rows and a Go seeder that generates Argon2id PIN hashes.

## 8. Go Workspace Notes

- `go.work` includes shared modules, all services, and `infra/scripts/seed_demo`
- `go.work.dev` is used inside service containers and excludes the seed CLI
- Compose mounts `go.work.dev` as `/workspace/go.work`

Reason: service containers only need service + shared modules for faster, simpler dev loops.

## 9. Service-Specific Behavior Highlights

Important documented rules from individual service READMEs:

- `ledger`: invariant `SUM(debits) == SUM(credits)` and idempotency key requirement for transaction posts
- `payments`: fraud dependency is fail-closed (decline if fraud is unavailable)
- `loans`: integrates both scoring and fraud in decision pipeline
- `agent`: offline-first with settlement tracking
- `branch`: region-scoped operations and four-eyes controls on sensitive actions
- `notification`: multilingual channel fanout (SMS/USSD/Voice/Telegram/Push)
- `adapter-hub`: `MESOB_ADAPTER_MODE` controls demo/live partner behavior

## 10. AI Services Details

### `ai/scoring`

- Features intentionally exclude protected attributes
- Fallback chain: ML model -> rules scorecard -> `INSUFFICIENT_HISTORY (422)`

Run locally:

```bash
cd ai/scoring
pip install -e .
uvicorn app.main:app --reload --port 9001
python -m pytest tests/
```

### `ai/fraud`

- Hybrid approach: deterministic rules + anomaly model
- Decisions: `ALLOW | REVIEW | BLOCK`
- Downstream integrations expected to fail closed

Run locally:

```bash
cd ai/fraud
pip install -e .
uvicorn app.main:app --reload --port 9002
python -m pytest tests/
```

## 11. Frontend and Channel Apps

### Mobile (`mobile/`)

- Yarn workspaces with `apps/user` and `apps/agent`
- Agent app supports offline queueing and batched sync using capture-time idempotency keys
- Languages: Amharic, Oromo, Tigrinya, English

Commands:

```bash
cd mobile
yarn install
yarn user
yarn agent
```

### Web Console (`web/console`)

- Next.js 14 app for super-admin and branch-officer workflows
- Uses bearer access token + refresh cookie pattern

Commands:

```bash
cd web/console
npm install
NEXT_PUBLIC_GATEWAY_URL=http://localhost:8000 npm run dev
```

## 12. OpenAPI and Swagger

Two doc entry points exist:

- `docs/openapi-combined.yaml` for one-page complete API surface
- Per-service specs in `services/*/api/openapi.yaml`

Compose `swagger-ui` mounts both and uses custom `docs/swagger-initializer.js`.

## 13. Common Commands Cheat Sheet

```bash
# Infra only
make dev

# Full platform
make all-up

# Go services
make build
make test
make lint
make clean

# Contracts/docs
make proto
make swagger-ui

# Demo data
make seed-demo
```

## 14. Troubleshooting

### Services cannot connect to Postgres

- Confirm compose infra is healthy: `docker compose ps`
- Host apps use `localhost:5433`; compose services use `postgres:5432`

### Fraud/scoring calls failing in loans/payments

- Verify both AI services are up on `9001` and `9002`
- Remember fail-closed behavior may intentionally decline operations

### Swagger page empty or wrong spec

- Confirm `docs/openapi-combined.yaml` exists and mounts correctly
- Check `docs/swagger-initializer.js` load path in compose

### Go module resolution issues in containers

- Ensure service containers mount `go.work.dev`
- Rebuild once if needed: `docker compose up -d --build`

## 15. Where to Go Next

- For service internals: read each `services/<name>/README.md`
- For AI internals: read `ai/scoring/README.md` and `ai/fraud/README.md`
- For architecture decisions: browse `docs/adr/`
- For operations: browse `docs/runbooks/`
