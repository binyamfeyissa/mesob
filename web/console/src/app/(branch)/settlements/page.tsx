"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { SettlementForm } from "@/features/settlements/SettlementForm";
import { Button } from "@/components/ui/Button";
import { Table } from "@/components/ui/Table";
import { Badge } from "@/components/ui/Badge";

interface Settlement {
  id: string;
  agent_id: string;
  amount_minor: number;
  status: string;
  confirmed_at: string;
}

async function fetchSettlements(): Promise<{ data: Settlement[] }> {
  return apiFetch<{ data: Settlement[] }>("/branch/settlements");
}

export default function SettlementsPage() {
  const [settlingAgentId, setSettlingAgentId] = useState<string | null>(null);
  const { data, isLoading, refetch } = useQuery({
    queryKey: ["settlements"],
    queryFn: fetchSettlements,
  });

  const columns = [
    { key: "agent_id", header: "Agent" },
    { key: "amount_minor", header: "Amount", render: (r: Settlement) => `${(r.amount_minor / 100).toFixed(2)} ETB` },
    { key: "status", header: "Status", render: (r: Settlement) => <Badge label={r.status} color="green" /> },
    { key: "confirmed_at", header: "Confirmed", render: (r: Settlement) => new Date(r.confirmed_at).toLocaleDateString() },
  ];

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Settlements</h1>
        <Button size="sm" onClick={() => setSettlingAgentId("NEW")}>New Settlement</Button>
      </div>

      {settlingAgentId && (
        <div className="mb-8 max-w-md bg-white rounded-xl border border-gray-200 p-5">
          <h2 className="font-semibold text-gray-900 mb-4">New Settlement</h2>
          <SettlementForm
            agentId={settlingAgentId}
            onSuccess={() => { setSettlingAgentId(null); refetch(); }}
          />
        </div>
      )}

      <Table
        columns={columns as any}
        data={(data?.data ?? []) as any[]}
        keyField={"id" as any}
        emptyMessage="No settlements yet"
        loading={isLoading}
      />
    </div>
  );
}
