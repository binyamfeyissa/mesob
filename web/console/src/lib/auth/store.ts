"use client";

import { create } from "zustand";
import { setAuth, clearAuth } from "@/lib/auth/index";
import { setAccessToken, clearAccessToken } from "@/lib/api/client";

const GATEWAY = process.env.NEXT_PUBLIC_GATEWAY_URL ?? "http://localhost:8000";

// Prevents concurrent hydration calls (e.g. AuthGuard mounting + manual call)
let _hydrating = false;
// Timestamp of the last successful session verification (ms)
let _lastVerifiedAt: number | null = null;
// Re-verify if the last check was longer than 5 minutes ago
const REVERIFY_THRESHOLD_MS = 5 * 60 * 1000;

export interface AuthUser {
  userId: string;
  role: string;
  kycTier: number;
}

interface AuthStore {
  isReady: boolean;
  isAuthenticated: boolean;
  user: AuthUser | null;
  hydrate: () => Promise<void>;
  login: (msisdn: string, pin: string) => Promise<AuthUser>;
  logout: () => Promise<void>;
}

export const useAuthStore = create<AuthStore>((set, get) => ({
  isReady: false,
  isAuthenticated: false,
  user: null,

  hydrate: async () => {
    if (_hydrating) return;

    // If the session was verified recently AND the store already has a valid
    // auth state, skip the round-trip to avoid a double-verify when navigating
    // from the root redirect (/) into a protected layout.
    const { isReady, isAuthenticated } = get();
    const recentlyVerified =
      _lastVerifiedAt !== null && Date.now() - _lastVerifiedAt < REVERIFY_THRESHOLD_MS;
    if (recentlyVerified && isReady && isAuthenticated) return;

    _hydrating = true;
    // Reset isReady so the spinner shows while the token refresh is in-flight.
    // This prevents a stale Zustand state (from a previous login in the same
    // browser tab) from bypassing auth after a server/Redis restart.
    set({ isReady: false });

    try {
      const res = await fetch(`${GATEWAY}/v1/identity/token/refresh`, {
        method: "POST",
        credentials: "include",
      });
      if (res.ok) {
        const body = await res.json();
        const d = body.data;
        setAccessToken(d.access_token);
        setAuth({ accessToken: d.access_token, role: d.role, kycTier: d.kyc_tier });
        _lastVerifiedAt = Date.now();
        set({
          isReady: true,
          isAuthenticated: true,
          user: { userId: d.user_id, role: d.role, kycTier: d.kyc_tier },
        });
        return;
      }
    } catch {
      // network error — treat as unauthenticated
    } finally {
      _hydrating = false;
    }
    clearAccessToken();
    clearAuth();
    set({ isReady: true, isAuthenticated: false, user: null });
  },

  login: async (msisdn, pin) => {
    const res = await fetch(`${GATEWAY}/v1/identity/login`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ msisdn, pin }),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body?.error?.message ?? "Login failed");
    }
    const body = await res.json();
    const d = body.data;
    setAccessToken(d.access_token);
    setAuth({ accessToken: d.access_token, role: d.role, kycTier: d.kyc_tier });
    const user: AuthUser = { userId: d.user_id, role: d.role, kycTier: d.kyc_tier };
    // Mark as recently verified so navigating to a protected page after login
    // does NOT trigger a redundant token refresh within the threshold window.
    _lastVerifiedAt = Date.now();
    set({ isReady: true, isAuthenticated: true, user });
    return user;
  },

  logout: async () => {
    await fetch(`${GATEWAY}/v1/identity/logout`, {
      method: "POST",
      credentials: "include",
    }).catch(() => {});
    _lastVerifiedAt = null;
    clearAccessToken();
    clearAuth();
    set({ isReady: true, isAuthenticated: false, user: null });
  },
}));
