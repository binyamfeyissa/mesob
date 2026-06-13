"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { KPICard } from "@/features/analytics/KPICard";
import { RecentAlerts } from "@/features/risk/RecentAlerts";
import { VolumeChart } from "@/features/analytics/VolumeChart";

interface Summary {
  users_active: number;
  txn_today: number;
  volume_today_minor: number;
  loans_active: number;
  open_alerts: number;
  float_health: string;
}

export default function DashboardPage() {
  const { data } = useQuery({
    queryKey: ["dashboard-summary"],
    queryFn: () => apiFetch<{ data: Summary }>("/admin/dashboard/summary"),
    refetchInterval: 30_000,
  });

  const s = data?.data;

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Dashboard</h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <KPICard title="Active Users" value={s?.users_active ?? "—"} />
        <KPICard title="Txns Today" value={s?.txn_today ?? "—"} />
        <KPICard
          title="Volume Today"
          value={
            s
              ? (s.volume_today_minor / 100).toLocaleString("en-ET", {
                  maximumFractionDigits: 2,
                })
              : "—"
          }
          suffix="ETB"
        />
        <KPICard title="Active Loans" value={s?.loans_active ?? "—"} />
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <VolumeChart />
        <RecentAlerts />
      </div>
    </div>
  );
}
