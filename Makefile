.PHONY: dev build test proto lint clean seed-demo swagger-ui

SERVICES := gateway identity ledger iqub iddir agent loans payments branch admin ussd telegram notification adapter-hub

dev:
	docker compose up -d postgres redis zookeeper kafka

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		(cd services/$$svc && go build ./...); \
	done

test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		(cd services/$$svc && go test ./internal/domain/... ./internal/app/...); \
	done

proto:
	./infra/scripts/gen-proto.sh

lint:
	@for svc in $(SERVICES); do \
		(cd services/$$svc && go vet ./...); \
	done

clean:
	@for svc in $(SERVICES); do \
		(cd services/$$svc && go clean ./...); \
	done

# Default DSN — matches docker-compose port 5433
IDENTITY_DB_URL ?= postgres://mesob:mesob@localhost:5433/mesob_identity?sslmode=disable

# Run SQL seed then Go seeder for demo accounts (PIN hashing requires Go)
seed-demo:
	@echo "--- Running SQL seed (regions, user rows) ---"
	PGPASSWORD=mesob psql -h localhost -p 5433 -U mesob -d mesob_identity -f infra/db/seeds/99-demo-accounts.sql
	@echo "--- Running Go seeder (Argon2id PIN hashes) ---"
	cd infra/scripts/seed_demo && go mod tidy && MESOB_IDENTITY_DB_URL="$(IDENTITY_DB_URL)" go run .

# Serve Swagger UI at http://localhost:8080 (no Docker required)
swagger-ui:
	@echo "Swagger UI → http://localhost:8080"
	npx --yes serve -l 8080 shared/openapi

.PHONY: all-up
all-up:
	docker compose up -d
