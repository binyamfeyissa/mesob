/**
 * Local read model schema for WatermelonDB (SQLite).
 * Mirrors server state for offline reads. Not a source of truth — server is authoritative.
 * Money is always stored as integer minor units (ETB cents).
 */

// Minimal type stubs — replace with actual @nozbe/watermelondb imports when installed
type TableSchema = {
  name: string;
  columns: { name: string; type: "string" | "number" | "boolean"; isOptional?: boolean }[];
};

export const accountsSchema: TableSchema = {
  name: "accounts",
  columns: [
    { name: "server_id", type: "string" },
    { name: "owner_type", type: "string" },
    { name: "owner_id", type: "string" },
    { name: "type", type: "string" },
    { name: "currency", type: "string" },
    { name: "balance_minor", type: "number" },   // int64 — ETB cents
    { name: "status", type: "string" },
    { name: "synced_at", type: "number" },
  ],
};

export const transactionsSchema: TableSchema = {
  name: "transactions",
  columns: [
    { name: "server_id", type: "string" },
    { name: "type", type: "string" },
    { name: "status", type: "string" },
    { name: "amount_minor", type: "number" },    // int64 — ETB cents
    { name: "direction", type: "string" },       // D | C from the perspective of the user's account
    { name: "channel", type: "string" },
    { name: "created_at", type: "number" },      // Unix timestamp ms
    { name: "idempotency_key", type: "string", isOptional: true },
  ],
};

export const iqubGroupsSchema: TableSchema = {
  name: "iqub_groups",
  columns: [
    { name: "server_id", type: "string" },
    { name: "name", type: "string" },
    { name: "cycle_minor", type: "number" },     // int64 — ETB cents
    { name: "frequency", type: "string" },
    { name: "status", type: "string" },
    { name: "cycle_number", type: "number" },
    { name: "cycle_paid", type: "number" },
    { name: "cycle_total", type: "number" },
    { name: "due_date", type: "string" },
    { name: "synced_at", type: "number" },
  ],
};

export const iddirGroupsSchema: TableSchema = {
  name: "iddir_groups",
  columns: [
    { name: "server_id", type: "string" },
    { name: "name", type: "string" },
    { name: "premium_minor", type: "number" },   // int64 — ETB cents
    { name: "benefit_minor", type: "number" },   // int64 — ETB cents
    { name: "frequency", type: "string" },
    { name: "coverage_status", type: "string" },
    { name: "synced_at", type: "number" },
  ],
};

export const loansSchema: TableSchema = {
  name: "loans",
  columns: [
    { name: "server_id", type: "string" },
    { name: "status", type: "string" },
    { name: "principal_minor", type: "number" }, // int64 — ETB cents
    { name: "outstanding_minor", type: "number" },
    { name: "due_date", type: "string" },
    { name: "synced_at", type: "number" },
  ],
};

export const pendingOpsSchema: TableSchema = {
  name: "pending_ops",
  columns: [
    { name: "idempotency_key", type: "string" }, // UUIDv7, generated at capture time
    { name: "op_type", type: "string" },         // CASH_IN | CASH_OUT | ONBOARD
    { name: "payload_json", type: "string" },
    { name: "captured_at", type: "number" },     // Unix timestamp ms
    { name: "status", type: "string" },          // PENDING | SYNCED | REJECTED
    { name: "attempts", type: "number" },
    { name: "rejection_reason", type: "string", isOptional: true },
  ],
};

export const appSchema = {
  version: 1,
  tables: [
    accountsSchema,
    transactionsSchema,
    iqubGroupsSchema,
    iddirGroupsSchema,
    loansSchema,
    pendingOpsSchema,
  ],
};
