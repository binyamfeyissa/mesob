// Access token lives in-process memory only — never persisted to storage.
// Refresh token is an httpOnly cookie set by the identity service — JS cannot read it.
// CSRF: on each refresh we read X-CSRF-Token response header and include it on next write.

let _accessToken: string | null = null;
let _role: string | null = null;
let _kycTier: number | null = null;
let _csrfToken: string | null = null;

export interface AuthState {
  accessToken: string;
  role: string;
  kycTier: number;
}

export function setAuth(state: AuthState): void {
  _accessToken = state.accessToken;
  _role = state.role;
  _kycTier = state.kycTier;
}

export function getAccessToken(): string | null {
  return _accessToken;
}

export function getRole(): string | null {
  return _role;
}

export function clearAuth(): void {
  _accessToken = null;
  _role = null;
  _kycTier = null;
  _csrfToken = null;
}

export function isAuthenticated(): boolean {
  return _accessToken !== null;
}

export function hasRole(...roles: string[]): boolean {
  return _role !== null && roles.includes(_role);
}

export function setCsrfToken(token: string): void {
  _csrfToken = token;
}

export function getCsrfToken(): string | null {
  return _csrfToken;
}
