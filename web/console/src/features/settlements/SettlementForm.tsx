"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

interface SettlementFormProps {
  agentId: string;
  onSuccess: () => void;
}

export function SettlementForm({ agentId, onSuccess }: SettlementFormProps) {
  const [amountETB, setAmountETB] = useState("");
  const [secondAuthoriser, setSecondAuthoriser] = useState("");
  const [error, setError] = useState("");
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => {
      const amount = parseFloat(amountETB);
      if (isNaN(amount) || amount <= 0) throw new Error("Enter a valid amount");
      if (!secondAuthoriser) throw new Error("Second authoriser ID is required");
      return apiFetch("/branch/settlements", {
        method: "POST",
        headers: { "Idempotency-Key": crypto.randomUUID() },
        body: JSON.stringify({
          agent_id: agentId,
          amount_minor: Math.round(amount * 100),
          second_authoriser_id: secondAuthoriser,
        }),
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["settlements"] });
      onSuccess();
    },
    onError: (e: Error) => setError(e.message),
  });

  return (
    <div className="space-y-4">
      <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-800">
        Four-eyes required: the second authoriser must be a different branch officer.
        <br />Error <code>SAME_AUTHORISER</code> will be returned if IDs match.
      </div>

      <Input
        label="Amount (ETB)"
        type="number"
        step="0.01"
        min="0"
        value={amountETB}
        onChange={(e) => setAmountETB(e.target.value)}
        placeholder="0.00"
      />

      <Input
        label="Second Authoriser ID"
        value={secondAuthoriser}
        onChange={(e) => setSecondAuthoriser(e.target.value)}
        placeholder="Officer UUID"
      />

      {error && <p className="text-red-500 text-sm">{error}</p>}

      <Button
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
      >
        {mutation.isPending ? "Submitting..." : "Confirm Settlement"}
      </Button>
    </div>
  );
}
