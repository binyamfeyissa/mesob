"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/lib/auth/store";

interface DemoAccount {
  label: string;
  msisdn: string;
  pin: string;
  role: string;
  color: string;
}

const DEMO_ACCOUNTS: DemoAccount[] = [
  { label: "Super Admin", msisdn: "+251911000001", pin: "111111", role: "SUPER_ADMIN", color: "bg-purple-600" },
  { label: "Admin",       msisdn: "+251911000002", pin: "111111", role: "ADMIN",       color: "bg-mesob-blue" },
  { label: "Branch Mgr",  msisdn: "+251911000003", pin: "111111", role: "BRANCH_MANAGER", color: "bg-teal-600" },
];

const ROLE_REDIRECT: Record<string, string> = {
  SUPER_ADMIN:    "/dashboard",
  ADMIN:          "/dashboard",
  BRANCH_MANAGER: "/kyc-review",
};

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login } = useAuthStore();

  const [msisdn, setMsisdn] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const doLogin = async (m: string, p: string) => {
    setLoading(true);
    setError("");
    try {
      const user = await login(m, p);
      const next = searchParams.get("next");
      router.replace(next ?? ROLE_REDIRECT[user.role] ?? "/dashboard");
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    doLogin(msisdn, pin);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-mesob-dark px-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-2xl shadow-2xl p-8">
          {/* Header */}
          <div className="mb-8 text-center">
            <h1 className="text-2xl font-bold text-mesob-dark">
              <span className="text-mesob-gold">Mesob</span> Wallet
            </h1>
            <p className="text-gray-400 text-sm mt-1">Admin Console</p>
          </div>

          {/* Login Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Phone Number (MSISDN)
              </label>
              <input
                type="tel"
                value={msisdn}
                onChange={(e) => setMsisdn(e.target.value)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2.5 text-sm focus:ring-2 focus:ring-mesob-blue focus:border-transparent outline-none"
                placeholder="+251911000002"
                required
                autoComplete="tel"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">PIN</label>
              <input
                type="password"
                value={pin}
                onChange={(e) => setPin(e.target.value)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2.5 text-sm focus:ring-2 focus:ring-mesob-blue focus:border-transparent outline-none"
                placeholder="••••••"
                maxLength={6}
                required
                autoComplete="current-password"
                inputMode="numeric"
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg px-3 py-2">
                <p className="text-red-600 text-sm">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-mesob-blue text-white rounded-lg py-2.5 text-sm font-semibold hover:opacity-90 transition disabled:opacity-50"
            >
              {loading ? "Signing in..." : "Sign In"}
            </button>
          </form>

          {/* Demo Quick-Login — dev/staging only */}
          {process.env.NODE_ENV !== "production" && (
            <div className="mt-8 pt-6 border-t border-gray-100">
              <p className="text-xs text-gray-400 text-center mb-3 uppercase tracking-wide">
                Dev Accounts — click to auto-fill
              </p>
              <div className="grid grid-cols-3 gap-2">
                {DEMO_ACCOUNTS.map((acc) => (
                  <button
                    key={acc.role}
                    type="button"
                    disabled={loading}
                    onClick={() => {
                      setMsisdn(acc.msisdn);
                      setPin(acc.pin);
                      doLogin(acc.msisdn, acc.pin);
                    }}
                    className={`${acc.color} text-white rounded-lg py-2 px-1 text-xs font-medium hover:opacity-90 transition disabled:opacity-50 flex flex-col items-center gap-0.5`}
                  >
                    <span>{acc.label}</span>
                    <span className="opacity-70 font-mono text-[10px]">PIN: {acc.pin}</span>
                  </button>
                ))}
              </div>
              <p className="text-[11px] text-gray-300 text-center mt-3">
                All dev accounts PIN: <span className="font-mono">111111</span>
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
