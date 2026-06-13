"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/auth/store";

interface AuthGuardProps {
  children: React.ReactNode;
  allowedRoles?: string[];
}

export function AuthGuard({ children, allowedRoles }: AuthGuardProps) {
  const { isReady, isAuthenticated, user, hydrate } = useAuthStore();
  const router = useRouter();

  // Always re-verify the session when this guard mounts.
  // hydrate() resets isReady:false first, so the spinner shows while the
  // token refresh is in-flight. This catches stale Zustand state after a
  // server/Redis restart even when the same browser tab is reused.
  useEffect(() => {
    hydrate();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // intentionally run once on mount only

  useEffect(() => {
    if (!isReady) return;
    if (!isAuthenticated) {
      router.replace("/login");
      return;
    }
    if (allowedRoles && user && !allowedRoles.includes(user.role)) {
      router.replace("/dashboard");
    }
  }, [isReady, isAuthenticated, user, router, allowedRoles]);

  if (!isReady) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-4 border-mesob-blue border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-gray-400">Verifying session...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) return null;
  if (allowedRoles && user && !allowedRoles.includes(user.role)) return null;

  return <>{children}</>;
}
