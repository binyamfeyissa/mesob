# Local Development Runbook

## Prerequisites
- Docker Desktop 4.x
- Go 1.25.5 (`go version`)
- Python 3.11 (`python3 --version`)
- Node.js 22.x (`node --version`)
- `make` available

## Start infrastructure
```bash
cd mesob
make dev
# Starts: PostgreSQL 16, Redis 7, Zookeeper, Kafka
# Waits for all health checks to pass
```

## Start individual Go services
```bash
# After infra is up:
cd services/identity && go run ./cmd/server
cd services/ledger && go run ./cmd/server

# Or build all:
make build
```

## Start Python AI services
```bash
cd ai/scoring
pip install -e .
uvicorn app.main:app --reload --port 9001

cd ai/fraud
pip install -e .
uvicorn app.main:app --reload --port 9002
```

## Start web console
```bash
cd web/console
npm install
NEXT_PUBLIC_GATEWAY_URL=http://localhost:8000 npm run dev
# Available at http://localhost:3000
```

## Verify everything is running
```bash
curl http://localhost:8000/health   # gateway
curl http://localhost:8001/health   # identity
curl http://localhost:8002/health   # ledger
curl http://localhost:9001/health   # scoring
curl http://localhost:9002/health   # fraud
```

## Database access
```bash
psql -h localhost -U mesob -d mesob_ledger
# Password: mesob (from docker-compose)
```

## Run tests
```bash
make test
# Or per-service:
cd services/identity && go test ./...
cd ai/scoring && python -m pytest tests/
```

## Proto generation
```bash
make proto
# Runs infra/scripts/gen-proto.sh
```
