#!/usr/bin/env bash
set -euo pipefail

PGURL="${MESOB_IDENTITY_DB_URL:-postgres://mesob:mesob@localhost:5432/mesob_identity}"

echo "Seeding regions..."
psql "$PGURL" <<'SQL'
INSERT INTO regions (id, name, code) VALUES
  (gen_random_uuid(), 'Addis Ababa', 'AA'),
  (gen_random_uuid(), 'Oromia', 'OR'),
  (gen_random_uuid(), 'Amhara', 'AM'),
  (gen_random_uuid(), 'Tigray', 'TI'),
  (gen_random_uuid(), 'SNNPR', 'SN')
ON CONFLICT DO NOTHING;
SQL

echo "Demo seed complete."
