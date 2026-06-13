import { operationQueue, QueuedOperation } from "./queue";

const GATEWAY_URL = process.env.EXPO_PUBLIC_GATEWAY_URL ?? "http://localhost:8000";

export interface SyncResult {
  appliedCount: number;
  rejectedCount: number;
  skippedCount: number; // ops not returned by server at all
}

export type OnItemDone = (
  id: string,
  outcome: "APPLIED" | "REJECTED",
  error?: string
) => void;

export async function syncPendingOperations(
  authToken: string,
  onItemDone?: OnItemDone
): Promise<SyncResult> {
  const pending = operationQueue.getPending();
  if (pending.length === 0) return { appliedCount: 0, rejectedCount: 0, skippedCount: 0 };

  const response = await fetch(`${GATEWAY_URL}/v1/agent/sync`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${authToken}`,
    },
    body: JSON.stringify({
      operations: pending.map((op) => ({
        idempotency_key: op.id,
        type: op.type,
        payload: op.payload,
      })),
    }),
  });

  if (!response.ok) {
    throw new Error(`Server returned ${response.status}`);
  }

  const result = await response.json();

  let appliedCount = 0;
  let rejectedCount = 0;

  for (const applied of result.data?.applied ?? []) {
    operationQueue.markSynced(applied.idempotency_key);
    onItemDone?.(applied.idempotency_key, "APPLIED");
    appliedCount++;
  }

  for (const rejected of result.data?.rejected ?? []) {
    // Keep rejected ops in the queue — agent needs to see the error
    operationQueue.markRejected(rejected.idempotency_key, rejected.error ?? "Rejected by server");
    onItemDone?.(rejected.idempotency_key, "REJECTED", rejected.error);
    rejectedCount++;
  }

  const respondedIds = new Set([
    ...(result.data?.applied ?? []).map((a: any) => a.idempotency_key),
    ...(result.data?.rejected ?? []).map((r: any) => r.idempotency_key),
  ]);
  const skippedCount = pending.filter((op) => !respondedIds.has(op.id)).length;

  return { appliedCount, rejectedCount, skippedCount };
}
