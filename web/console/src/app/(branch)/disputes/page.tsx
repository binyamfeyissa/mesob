"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Badge } from "@/components/ui/Badge";
import { Table } from "@/components/ui/Table";

interface Dispute {
  id: string;
  transaction_id: string;
  raised_by: string;
  reason: string;
  resolution: string | null;
  created_at: string;
}

async function fetchDisputes(): Promise<{ data: Dispute[] }> {
  return apiFetch<{ data: Dispute[] }>("/branch/disputes");
}

export default function DisputesPage() {
  const qc = useQueryClient();
  const [selected, setSelected] = useState<Dispute | null>(null);
  const [secondAuth, setSecondAuth] = useState("");
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["disputes"],
    queryFn: fetchDisputes,
    refetchInterval: 30_000,
  });

  const mutation = useMutation({
    mutationFn: (resolution: "REFUND" | "DENY") => {
      if (!secondAuth) throw new Error("Second authoriser is required");
      return apiFetch(`/branch/disputes/${selected!.id}/resolve`, {
        method: "POST",
        body: JSON.stringify({ resolution, second_authoriser_id: secondAuth, reason }),
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["disputes"] });
      setSelected(null); setSecondAuth(""); setReason(""); setError("");
    },
    onError: (e: Error) => setError(e.message),
  });

  if (selected) {
    return (
      <div className="p-6 max-w-lg">
        <button onClick={() => setSelected(null)} className="text-sm text-mesob-blue mb-4 hover:underline">
          ← Back
        </button>
        <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-4">
          <h2 className="font-semibold text-gray-900">Resolve Dispute</h2>
          <div className="text-sm text-gray-600 space-y-1">
            <p>Txn: <span className="font-mono">{selected.transaction_id}</span></p>
            <p>Reason: {selected.reason}</p>
          </div>
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-800">
            Four-eyes required. REFUND triggers a compensating Ledger reversal.
          </div>
          <Input label="Second Authoriser ID" value={secondAuth} onChange={(e) => setSecondAuth(e.target.value)} placeholder="Officer UUID" />
          <Input label="Resolution Note" value={reason} onChange={(e) => setReason(e.target.value)} placeholder="Reason for decision" />
          {error && <p className="text-red-500 text-sm">{error}</p>}
          <div className="flex gap-2">
            <Button onClick={() => mutation.mutate("REFUND")} disabled={mutation.isPending}>Refund</Button>
            <Button variant="danger" onClick={() => mutation.mutate("DENY")} disabled={mutation.isPending}>Deny</Button>
          </div>
        </div>
      </div>
    );
  }

  const columns = [
    { key: "transaction_id", header: "Transaction" },
    { key: "reason", header: "Reason" },
    { key: "created_at", header: "Raised", render: (r: Dispute) => new Date(r.created_at).toLocaleDateString() },
    {
      key: "actions", header: "",
      render: (r: Dispute) => (
        <Button size="sm" variant="secondary" onClick={() => setSelected(r)}>Review</Button>
      ),
    },
  ];

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Disputes</h1>
      <Table columns={columns as any} data={(data?.data ?? []) as any[]} keyField={"id" as any} emptyMessage="No open disputes" loading={isLoading} />
    </div>
  );
}
