"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/auth/store";

const ADMIN_NAV = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/risk", label: "Risk & Fraud" },
  { href: "/audit", label: "Audit Log" },
  { href: "/config", label: "Config & Flags" },
  { href: "/live", label: "Live Stream" },
];

const BRANCH_NAV = [
  { href: "/kyc-review", label: "KYC Review" },
  { href: "/settlements", label: "Settlements" },
  { href: "/disputes", label: "Disputes" },
  { href: "/reconciliation", label: "Reconciliation" },
];

const ROLE_LABEL: Record<string, string> = {
  SUPER_ADMIN:    "Super Admin",
  ADMIN:          "Admin",
  BRANCH_MANAGER: "Branch Mgr",
};

const ROLE_COLOR: Record<string, string> = {
  SUPER_ADMIN:    "bg-purple-600",
  ADMIN:          "bg-mesob-blue",
  BRANCH_MANAGER: "bg-teal-600",
};

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();

  const role = user?.role ?? "";
  const isBranchOnly = role === "BRANCH_MANAGER";
  const navItems = isBranchOnly ? BRANCH_NAV : [...ADMIN_NAV, ...BRANCH_NAV];

  const handleLogout = async () => {
    await logout();
    router.replace("/login");
  };

  return (
    <aside className="w-56 bg-mesob-dark text-white flex flex-col">
      {/* Logo */}
      <div className="px-4 py-5 border-b border-white/10">
        <span className="text-lg font-bold text-mesob-gold">Mesob</span>
        <span className="text-lg font-bold text-white"> Wallet</span>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center px-3 py-2 rounded-lg text-sm font-medium transition ${
              pathname === item.href
                ? "bg-mesob-blue text-white"
                : "text-gray-300 hover:bg-white/10"
            }`}
          >
            {item.label}
          </Link>
        ))}
      </nav>

      {/* User footer */}
      <div className="px-3 py-4 border-t border-white/10 space-y-3">
        {user && (
          <div className="flex items-center gap-2">
            <span
              className={`${ROLE_COLOR[role] ?? "bg-gray-600"} text-white text-xs font-semibold px-2 py-0.5 rounded-full`}
            >
              {ROLE_LABEL[role] ?? role}
            </span>
            <span className="text-gray-400 text-xs font-mono truncate">{user.userId.slice(-8)}</span>
          </div>
        )}
        <button
          onClick={handleLogout}
          className="w-full text-left px-3 py-2 rounded-lg text-sm text-gray-300 hover:bg-white/10 transition"
        >
          Sign Out
        </button>
      </div>
    </aside>
  );
}
