"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/auth/store";
import { apiFetch } from "@/lib/api/client";

interface UserProfile {
  user_id: string;
  msisdn: string;
  role: string;
  kyc_tier: number;
  status: string;
}

const ROLE_LABEL: Record<string, string> = {
  SUPER_ADMIN: "Super Admin",
  ADMIN: "Admin",
  BRANCH_MANAGER: "Branch Mgr",
};

const ROLE_COLOR: Record<string, string> = {
  SUPER_ADMIN: "bg-purple-100 text-purple-800",
  ADMIN: "bg-blue-100 text-blue-800",
  BRANCH_MANAGER: "bg-teal-100 text-teal-800",
};

export function UserHeader() {
  const { isReady, isAuthenticated, user, logout } = useAuthStore();
  const router = useRouter();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);

  // Redirect to login if auth check completes and user is not authenticated
  useEffect(() => {
    if (!isReady) return;
    if (!isAuthenticated) {
      router.replace("/login");
    }
  }, [isReady, isAuthenticated, router]);

  // Fetch the full user profile (MSISDN, status, etc.) once authenticated
  useEffect(() => {
    if (!isReady || !isAuthenticated) return;
    setProfileLoading(true);
    apiFetch<{ data: UserProfile }>("/identity/me")
      .then((res) => setProfile(res.data))
      .catch(() => {})
      .finally(() => setProfileLoading(false));
  }, [isReady, isAuthenticated]);

  // While session is being verified, show a header skeleton
  if (!isReady) {
    return (
      <div className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between shadow-sm">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gray-200 animate-pulse" />
          <div className="space-y-1.5">
            <div className="h-4 w-24 bg-gray-200 rounded animate-pulse" />
            <div className="h-3 w-32 bg-gray-100 rounded animate-pulse" />
          </div>
        </div>
      </div>
    );
  }

  // If not authenticated (redirect in progress), keep showing skeleton to avoid flash
  if (!isAuthenticated || !user) {
    return (
      <div className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between shadow-sm">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gray-200 animate-pulse" />
          <div className="space-y-1.5">
            <div className="h-4 w-24 bg-gray-200 rounded animate-pulse" />
          </div>
        </div>
      </div>
    );
  }

  const roleLabel = ROLE_LABEL[user.role] ?? user.role;
  const roleColor = ROLE_COLOR[user.role] ?? "bg-gray-100 text-gray-700";
  const initial = roleLabel.charAt(0).toUpperCase();

  return (
    <div className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between shadow-sm">
      <div className="flex items-center gap-3">
        {/* Avatar */}
        <div className="w-8 h-8 rounded-full bg-mesob-blue flex items-center justify-center text-white text-xs font-bold shrink-0">
          {initial}
        </div>

        {/* Identity details */}
        <div>
          <div className="flex items-center gap-2 flex-wrap">
            <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${roleColor}`}>
              {roleLabel}
            </span>
            {profileLoading ? (
              <div className="h-3.5 w-28 bg-gray-100 rounded animate-pulse" />
            ) : profile ? (
              <span className="text-sm text-gray-700 font-mono">{profile.msisdn}</span>
            ) : (
              <span className="text-xs text-gray-400 font-mono">{user.userId.slice(-8)}</span>
            )}
          </div>
          {profile && (
            <p className="text-xs text-gray-400 mt-0.5">
              KYC Tier {profile.kyc_tier} · {profile.status}
            </p>
          )}
        </div>
      </div>

      <button
        onClick={async () => {
          await logout();
          router.replace("/login");
        }}
        className="text-xs text-gray-400 hover:text-red-600 transition px-3 py-1.5 rounded-lg hover:bg-red-50"
      >
        Sign out
      </button>
    </div>
  );
}
