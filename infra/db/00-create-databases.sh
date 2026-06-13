#!/usr/bin/env bash
# Creates all service databases. Docker auto-creates mesob_identity (POSTGRES_DB),
# so we skip it and create only the remaining ones idempotently.
set -e

create_db() {
  local db="$1"
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" --dbname postgres -tc \
    "SELECT 1 FROM pg_database WHERE datname = '$db'" \
    | grep -q 1 || psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" --dbname postgres \
    -c "CREATE DATABASE $db OWNER $POSTGRES_USER;"
}

create_db mesob_ledger
create_db mesob_iqub
create_db mesob_iddir
create_db mesob_agent
create_db mesob_loans
create_db mesob_payments
create_db mesob_branch
create_db mesob_admin
create_db mesob_notification
create_db mesob_scoring
create_db mesob_fraud
create_db mesob_adapters
