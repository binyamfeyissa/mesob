/**
 * Persisted outbound operation queue. Survives app restarts via AsyncStorage.
 * Each operation carries a UUIDv7 key generated at capture time.
 */
import AsyncStorage from "@react-native-async-storage/async-storage";

const STORAGE_KEY = "agent_op_queue";

export interface QueuedOperation {
  id: string;              // UUIDv7 generated at capture
  type: "CASH_IN" | "CASH_OUT" | "ONBOARD";
  payload: Record<string, unknown>;
  capturedAt: string;      // ISO8601
  status: "PENDING" | "SYNCED" | "REJECTED";
  attempts: number;
  syncError?: string;      // server rejection reason, kept for display
}

export class OperationQueue {
  private operations: QueuedOperation[] = [];
  private hydrated = false;

  /** Load persisted queue from AsyncStorage. Must be called at app startup. */
  async hydrate(): Promise<void> {
    if (this.hydrated) return;
    try {
      const raw = await AsyncStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed: QueuedOperation[] = JSON.parse(raw);
        // Keep only ops that are not permanently resolved
        this.operations = parsed.filter((op) => op.status !== "SYNCED");
      }
    } catch {
      // Storage unavailable — start with empty queue
    }
    this.hydrated = true;
  }

  enqueue(op: Omit<QueuedOperation, "status" | "attempts">): void {
    this.operations.push({ ...op, status: "PENDING", attempts: 0 });
    this.persist();
  }

  getPending(): QueuedOperation[] {
    return this.operations.filter((op) => op.status === "PENDING");
  }

  /** All ops that are not permanently resolved (PENDING + REJECTED). */
  getAll(): QueuedOperation[] {
    return this.operations.filter((op) => op.status !== "SYNCED");
  }

  markSynced(id: string): void {
    const op = this.operations.find((o) => o.id === id);
    if (op) { op.status = "SYNCED"; this.persist(); }
  }

  markRejected(id: string, error?: string): void {
    const op = this.operations.find((o) => o.id === id);
    if (op) {
      op.status = "REJECTED";
      op.syncError = error;
      this.persist();
    }
  }

  /** Reset a rejected op back to PENDING so it can be retried. */
  resetToPending(id: string): void {
    const op = this.operations.find((o) => o.id === id);
    if (op && op.status === "REJECTED") {
      op.status = "PENDING";
      op.syncError = undefined;
      this.persist();
    }
  }

  /** Permanently remove a rejected op (agent manually dismisses it). */
  dismiss(id: string): void {
    this.operations = this.operations.filter((o) => o.id !== id);
    this.persist();
  }

  private persist(): void {
    AsyncStorage.setItem(STORAGE_KEY, JSON.stringify(this.operations)).catch(() => {
      // Non-fatal: in-memory state is still correct
    });
  }
}

export const operationQueue = new OperationQueue();
