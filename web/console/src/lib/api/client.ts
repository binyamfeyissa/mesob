const GATEWAY_URL = process.env.NEXT_PUBLIC_GATEWAY_URL ?? "http://localhost:8000";

let accessToken: string | null = null;

export function setAccessToken(token: string) {
  accessToken = token;
}

export function clearAccessToken() {
  accessToken = null;
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const res = await fetch(`${GATEWAY_URL}/v1${path}`, {
    ...options,
    headers,
    credentials: "include", // send refresh token cookie
  });

  if (res.status === 401 && accessToken) {
    // Attempt transparent token refresh (once)
    const refreshed = await tryRefresh();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${accessToken}`;
      const retry = await fetch(`${GATEWAY_URL}/v1${path}`, {
        ...options,
        headers,
        credentials: "include",
      });
      if (!retry.ok) throw new Error(`API error ${retry.status}`);
      return retry.json();
    }
    // Refresh failed → force re-login
    clearAccessToken();
    window.location.href = "/login";
    throw new Error("Session expired");
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.code ?? `API error ${res.status}`);
  }

  return res.json();
}

async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${GATEWAY_URL}/v1/identity/token/refresh`, {
      method: "POST",
      credentials: "include",
    });
    if (!res.ok) return false;
    const data = await res.json();
    setAccessToken(data.data.access_token);
    return true;
  } catch {
    return false;
  }
}
