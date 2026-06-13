/**
 * UUIDv7 idempotency keys are generated at CAPTURE TIME, not at sync time.
 * This ensures that offline operations replayed during sync are safe.
 */
export function generateCaptureKey(): string {
  // UUIDv7: time-ordered UUID. In production use a proper UUIDv7 library.
  // Approximation: timestamp prefix + random suffix
  const now = Date.now();
  const timestamp = now.toString(16).padStart(12, "0");
  const random = Math.random().toString(16).slice(2).padStart(20, "0");
  return `${timestamp.slice(0, 8)}-${timestamp.slice(8)}-7${random.slice(0, 3)}-${random.slice(3, 7)}-${random.slice(7, 19)}`;
}
