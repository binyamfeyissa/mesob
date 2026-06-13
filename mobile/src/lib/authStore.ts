import { create } from "zustand";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { setAccessToken } from "../api/client";

const GATEWAY = process.env.EXPO_PUBLIC_GATEWAY_URL ?? "http://localhost:8000";

export interface AuthUser {
  userId: string;
  role: string;
  kycTier: number;
  msisdn: string;
  walletAccountId: string;
}

interface AuthStore {
  isReady: boolean;
  isAuthenticated: boolean;
  user: AuthUser | null;
  hydrate: () => Promise<void>;
  setSession: (data: {
    accessToken: string;
    refreshToken: string;
    role: string;
    userId: string;
    kycTier: number;
    msisdn: string;
    walletAccountId?: string;
  }) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  isReady: false,
  isAuthenticated: false,
  user: null,

  hydrate: async () => {
    const rt = await AsyncStorage.getItem("refresh_token");
    if (!rt) {
      set({ isReady: true });
      return;
    }
    try {
      const res = await fetch(`${GATEWAY}/v1/identity/token/refresh`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Refresh-Token": rt,
        },
      });
      if (res.ok) {
        const body = await res.json();
        const d = body.data;
        setAccessToken(d.access_token);
        await AsyncStorage.setItem("refresh_token", d.refresh_token);
        const msisdn = (await AsyncStorage.getItem("msisdn")) ?? "";
        set({
          isReady: true,
          isAuthenticated: true,
          user: { userId: d.user_id, role: d.role, kycTier: d.kyc_tier, msisdn, walletAccountId: d.wallet_account_id ?? "" },
        });
        return;
      }
    } catch {
      // network error — will remain unauthenticated
    }
    await AsyncStorage.removeItem("refresh_token");
    set({ isReady: true, isAuthenticated: false, user: null });
  },

  setSession: ({ accessToken, refreshToken, role, userId, kycTier, msisdn, walletAccountId = "" }) => {
    setAccessToken(accessToken);
    AsyncStorage.setItem("refresh_token", refreshToken);
    AsyncStorage.setItem("msisdn", msisdn);
    set({ isAuthenticated: true, user: { userId, role, kycTier, msisdn, walletAccountId } });
  },

  logout: () => {
    AsyncStorage.removeItem("refresh_token");
    AsyncStorage.removeItem("msisdn");
    setAccessToken("");
    set({ isAuthenticated: false, user: null });
  },
}));
