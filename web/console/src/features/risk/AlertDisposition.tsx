"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";

interface Alert {
  alert_id: string;
  severity: string;
  rule: string;
  transaction_id: string;
  status: string;
}

interface AlertDispositionProps {
  alert: Alert;
  onDone: () => void;
}

type Disposition = "CLEAR" | "CONFIRM" | "ESCALATE_SAR";

export function AlertDisposition({ alert, onDone }: AlertDispositionProps) {
  const [disposition, setDisposition] = useState<Disposition | null>(null);
  const [note, setNote] = useState("");
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (d: Disposition) =>
      apiFetch(`/fraud/alerts/${alert.alert_id}/disposition`, {
        method: "POST",
        body: JSON.stringify({ disposition: d, note }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fraud-alerts"] });
      onDone();
    },
  });

  const severityColor = alert.severity === "HIGH" ? "red" : alert.severity === "MEDIUM" ? "yellow" : "gray";

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <p className="font-semibold text-gray-900 font-mono text-sm">{alert.rule}</p>
          <p className="text-xs text-gray-400 mt-1">Txn: {alert.transaction_id}</p>
        </div>
        <Badge label={alert.severity} color={severityColor as any} />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-600 mb-1">Note (required for SAR)</label>
        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={2}
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-mesob-blue"
          placeholder="Add investigation notes..."
        />
      </div>

      <div className="flex gap-2">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => mutation.mutate("CLEAR")}
          disabled={mutation.isPending}
        >
          Clear
        </Button>
        <Button
          size="sm"
          onClick={() => mutation.mutate("CONFIRM")}
          disabled={mutation.isPending}
        >
          Confirm Fraud
        </Button>
        <Button
          variant="danger"
          size="sm"
          onClick={() => {
            if (!note) { window.alert("A note is required for SAR escalation"); return; }
            mutation.mutate("ESCALATE_SAR");
          }}
          disabled={mutation.isPending}
        >
          Escalate SAR
        </Button>
      </div>

      {mutation.isError && (
        <p className="text-red-500 text-xs">{String(mutation.error)}</p>
      )}
    </div>
  );
}
