"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Badge } from "@/components/ui/Badge";

interface ReconciliationReport {
  region_id: string;
  date: string;
  ledger_minor: number;
  counted_minor: number;
  variance_minor: number;
  status: "BALANCED" | "VARIANCE";
}

export default function ReconciliationPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["reconciliation"],
    queryFn: () => apiFetch<{ data: ReconciliationReport }>("/branch/reconciliation"),
    refetchInterval: 60_000,
  });

  const r = data?.data;

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Reconciliation</h1>

      {isLoading && <p className="text-gray-400 text-sm">Loading...</p>}

      {r && (
        <div className="max-w-md space-y-4">
          <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">Date</span>
              <span className="font-medium text-gray-900">{r.date}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">Ledger Balance</span>
              <span className="font-mono font-medium">{(r.ledger_minor / 100).toFixed(2)} ETB</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">Physical Count</span>
              <span className="font-mono font-medium">{(r.counted_minor / 100).toFixed(2)} ETB</span>
            </div>
            <div className="flex items-center justify-between border-t pt-3">
              <span className="text-sm font-semibold text-gray-700">Variance</span>
              <span className={`font-mono font-bold ${r.variance_minor !== 0 ? "text-red-600" : "text-green-600"}`}>
                {(r.variance_minor / 100).toFixed(2)} ETB
              </span>
            </div>
            <div className="flex justify-end">
              <Badge
                label={r.status}
                color={r.status === "BALANCED" ? "green" : "red"}
              />
            </div>
          </div>

          {r.status === "VARIANCE" && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-sm text-red-700">
              Variance detected. Escalate to supervisor and file a dispute if needed.
            </div>
          )}
        </div>
      )}
    </div>
  );
}
