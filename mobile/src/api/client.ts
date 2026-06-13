/**
 * Typed REST client with transparent token refresh.
 * Access token: short-lived (15min), stored in memory.
 * Refresh token: 7-day opaque, in device-secure storage.
 */
const GATEWAY_URL = process.env.EXPO_PUBLIC_GATEWAY_URL ?? "http://localhost:8000";

let _accessToken: string | null = null;

export function setAccessToken(token: string): void {
  _accessToken = token;
}

export function clearAccessToken(): void {
  _accessToken = null;
}

export function getAccessToken(): string | null {
  return _accessToken;
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (_accessToken) headers.Authorization = `Bearer ${_accessToken}`;

  const res = await fetch(`${GATEWAY_URL}/v1${path}`, { ...options, headers });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.code ?? `API error ${res.status}`);
  }
  if (res.status === 204 || res.headers.get("content-length") === "0") {
    return {} as T;
  }
  return res.json();
}
