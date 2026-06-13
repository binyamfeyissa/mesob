"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Table } from "@/components/ui/Table";

interface KYCRequest {
  user_id: string;
  msisdn: string;
  current_tier: number;
  requested_tier: number;
  submitted_at: string;
  status: string;
}

export function KYCReviewList() {
  const queryClient = useQueryClient();
  const [note, setNote] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["kyc-pending"],
    queryFn: () => apiFetch<{ data: KYCRequest[] }>("/branch/kyc/queue"),
    refetchInterval: 30_000,
  });

  const mutation = useMutation({
    mutationFn: ({ userId, decision, tier }: { userId: string; decision: "APPROVE" | "REJECT"; tier: number }) =>
      apiFetch(`/branch/kyc/${userId}/review`, {
        method: "POST",
        body: JSON.stringify({ decision, target_tier: tier, note }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kyc-pending"] });
      setNote("");
    },
  });

  const columns = [
    { key: "msisdn", header: "MSISDN" },
    {
      key: "current_tier",
      header: "Current Tier",
      render: (r: KYCRequest) => <Badge label={`Tier ${r.current_tier}`} color="gray" />,
    },
    {
      key: "requested_tier",
      header: "Requested Tier",
      render: (r: KYCRequest) => <Badge label={`Tier ${r.requested_tier}`} color="blue" />,
    },
    {
      key: "submitted_at",
      header: "Submitted",
      render: (r: KYCRequest) => new Date(r.submitted_at).toLocaleString(),
    },
    {
      key: "actions",
      header: "Actions",
      render: (r: KYCRequest) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={() => mutation.mutate({ userId: r.user_id, decision: "APPROVE", tier: r.requested_tier })}
            disabled={mutation.isPending}
          >
            Approve
          </Button>
          <Button
            size="sm"
            variant="danger"
            onClick={() => mutation.mutate({ userId: r.user_id, decision: "REJECT", tier: r.requested_tier })}
            disabled={mutation.isPending}
          >
            Reject
          </Button>
        </div>
      ),
    },
  ] as any[];

  return (
    <Table
      columns={columns}
      data={(data?.data ?? []) as any[]}
      keyField={"user_id" as any}
      emptyMessage="No pending KYC reviews"
      loading={isLoading}
    />
  );
}
