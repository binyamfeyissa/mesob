"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Table } from "@/components/ui/Table";
import { Input } from "@/components/ui/Input";

interface AuditEntry {
  id: string;
  actor_id: string;
  actor_role: string;
  action: string;
  target: string;
  channel: string;
  ip: string;
  created_at: string;
}

export default function AuditPage() {
  const [actorId, setActorId] = useState("");
  const [action, setAction] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const params = new URLSearchParams();
  if (actorId) params.set("actor_id", actorId);
  if (action) params.set("action", action);
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  params.set("limit", "50");

  const { data, isLoading } = useQuery({
    queryKey: ["audit", actorId, action, from, to],
    queryFn: () => apiFetch<{ data: AuditEntry[] }>(`/admin/audit?${params}`),
    refetchInterval: 30_000,
  });

  const columns = [
    { key: "created_at", header: "Time", render: (r: AuditEntry) => new Date(r.created_at).toLocaleString() },
    { key: "actor_role", header: "Role" },
    { key: "action", header: "Action" },
    { key: "target", header: "Target" },
    { key: "channel", header: "Channel" },
    { key: "ip", header: "IP" },
  ];

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Audit Log</h1>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
        <Input label="Actor ID" value={actorId} onChange={(e) => setActorId(e.target.value)} placeholder="UUID" />
        <Input label="Action" value={action} onChange={(e) => setAction(e.target.value)} placeholder="e.g. KYC_APPROVE" />
        <Input label="From" type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        <Input label="To" type="date" value={to} onChange={(e) => setTo(e.target.value)} />
      </div>

      <Table
        columns={columns as any}
        data={(data?.data ?? []) as any[]}
        keyField={"id" as any}
        emptyMessage="No audit entries"
        loading={isLoading}
      />
    </div>
  );
}
