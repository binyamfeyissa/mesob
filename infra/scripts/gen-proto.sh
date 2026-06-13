#!/usr/bin/env bash
set -euo pipefail

PROTO_DIR="$(dirname "$0")/../../shared/proto"
SERVICES=(ledger fraud scoring notify adapters)

for svc in "${SERVICES[@]}"; do
  echo "Generating proto: $svc"
  protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_DIR/$svc.proto"
done
echo "Proto generation complete."
