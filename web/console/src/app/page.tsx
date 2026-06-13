"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/auth/store";

const ROLE_HOME: Record<string, string> = {
  SUPER_ADMIN:    "/dashboard",
  ADMIN:          "/dashboard",
  BRANCH_MANAGER: "/kyc-review",
};

export default function Home() {
  const { isReady, isAuthenticated, user, hydrate } = useAuthStore();
  const router = useRouter();

  // The root page is not wrapped in an AuthGuard layout, so it must trigger
  // hydration itself.
  useEffect(() => {
    hydrate();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!isReady) return;
    if (!isAuthenticated) {
      router.replace("/login");
      return;
    }
    const dest = user ? (ROLE_HOME[user.role] ?? "/dashboard") : "/dashboard";
    router.replace(dest);
  }, [isReady, isAuthenticated, user, router]);

  // Show a full-page spinner while the session is being verified
  if (!isReady) {
    return (
      <div className="flex h-screen items-center justify-center bg-mesob-dark">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-4 border-mesob-gold border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-gray-400">Loading...</p>
        </div>
      </div>
    );
  }

  return null;
}
