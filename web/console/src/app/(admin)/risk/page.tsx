"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Badge } from "@/components/ui/Badge";
import { AlertDisposition } from "@/features/risk/AlertDisposition";

interface Alert {
  alert_id: string;
  severity: string;
  rule: string;
  transaction_id: string;
  status: string;
}

async function fetchAlerts(): Promise<{ data: Alert[] }> {
  return apiFetch("/admin/fraud/alerts");
}

export default function RiskPage() {
  const [selected, setSelected] = useState<Alert | null>(null);
  const { data, isLoading, refetch } = useQuery({
    queryKey: ["admin-risk-alerts"],
    queryFn: fetchAlerts,
    refetchInterval: 15_000,
  });

  const alerts = data?.data ?? [];

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Risk & Fraud Alerts</h1>

      {selected ? (
        <div className="max-w-lg">
          <button
            onClick={() => setSelected(null)}
            className="text-sm text-mesob-blue mb-4 hover:underline"
          >
            ← Back to alerts
          </button>
          <AlertDisposition alert={selected} onDone={() => { setSelected(null); refetch(); }} />
        </div>
      ) : (
        <div className="space-y-3">
          {isLoading && <p className="text-gray-400 text-sm">Loading...</p>}
          {alerts.length === 0 && !isLoading && (
            <p className="text-gray-400 text-sm">No open alerts.</p>
          )}
          {alerts.map((a) => (
            <button
              key={a.alert_id}
              onClick={() => setSelected(a)}
              className="w-full text-left bg-white rounded-xl border border-gray-200 p-4 hover:border-mesob-blue transition"
            >
              <div className="flex items-center justify-between">
                <span className="font-mono text-sm font-medium text-gray-800">{a.rule}</span>
                <Badge
                  label={a.severity}
                  color={a.severity === "HIGH" ? "red" : a.severity === "MEDIUM" ? "yellow" : "gray"}
                />
              </div>
              <p className="text-xs text-gray-400 mt-1">Txn: {a.transaction_id}</p>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
