"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Badge } from "@/components/ui/Badge";

interface Alert {
  alert_id: string;
  severity: string;
  rule: string;
  transaction_id: string;
  status: string;
}

export function RecentAlerts() {
  const { data, isLoading } = useQuery({
    queryKey: ["fraud-alerts"],
    queryFn: () => apiFetch<{ data: Alert[] }>("/admin/fraud/alerts"),
    refetchInterval: 30_000,
  });

  const alerts = data?.data ?? [];

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-5">
      <h2 className="text-sm font-semibold text-gray-700 mb-4">Recent Fraud Alerts</h2>
      {isLoading ? (
        <p className="text-sm text-gray-400">Loading...</p>
      ) : alerts.length === 0 ? (
        <p className="text-sm text-gray-400">No open alerts.</p>
      ) : (
        <ul className="space-y-3">
          {alerts.map((a) => (
            <li key={a.alert_id} className="flex items-center justify-between text-sm">
              <span className="text-gray-700 font-mono">{a.rule}</span>
              <Badge label={a.severity} color={a.severity === "HIGH" ? "red" : "yellow"} />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
