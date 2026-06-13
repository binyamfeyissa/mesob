"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Badge } from "@/components/ui/Badge";

interface FeatureFlag {
  name: string;
  enabled: boolean;
  rollout_pct: number;
}

interface ConfigItem {
  key: string;
  value: unknown;
  version: number;
}

export default function ConfigPage() {
  const qc = useQueryClient();
  const [configKey, setConfigKey] = useState("");
  const [configValue, setConfigValue] = useState("");
  const [secondAuth, setSecondAuth] = useState("");
  const [reason, setReason] = useState("");
  const [configError, setConfigError] = useState("");

  const { data: flags, isLoading: flagsLoading } = useQuery({
    queryKey: ["flags"],
    queryFn: () => apiFetch<{ data: FeatureFlag[] }>("/admin/flags"),
  });

  const flagMutation = useMutation({
    mutationFn: ({ flag, enabled, rollout_pct }: { flag: string; enabled: boolean; rollout_pct: number }) =>
      apiFetch(`/admin/flags/${flag}`, {
        method: "PATCH",
        body: JSON.stringify({ enabled, rollout_pct }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["flags"] }),
  });

  const configMutation = useMutation({
    mutationFn: () => {
      if (!configKey || !secondAuth || !reason) {
        throw new Error("Key, second authoriser, and reason are all required");
      }
      let parsed: unknown;
      try { parsed = JSON.parse(configValue); } catch { parsed = configValue; }
      return apiFetch(`/admin/config/${configKey}`, {
        method: "PUT",
        body: JSON.stringify({ value: parsed, second_authoriser_id: secondAuth, reason }),
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["config"] });
      setConfigKey(""); setConfigValue(""); setSecondAuth(""); setReason(""); setConfigError("");
    },
    onError: (e: Error) => setConfigError(e.message),
  });

  return (
    <div className="p-6 space-y-8">
      <div>
        <h1 className="text-xl font-semibold text-gray-900 mb-4">Feature Flags</h1>
        {flagsLoading ? (
          <p className="text-gray-400 text-sm">Loading...</p>
        ) : (
          <div className="space-y-2">
            {(flags?.data ?? []).map((f) => (
              <div key={f.name} className="flex items-center justify-between bg-white border border-gray-200 rounded-lg px-4 py-3">
                <div>
                  <span className="font-mono text-sm font-medium text-gray-800">{f.name}</span>
                  <span className="ml-3 text-xs text-gray-400">Rollout: {f.rollout_pct}%</span>
                </div>
                <div className="flex items-center gap-3">
                  <Badge label={f.enabled ? "ON" : "OFF"} color={f.enabled ? "green" : "gray"} />
                  <Button
                    size="sm"
                    variant={f.enabled ? "danger" : "primary"}
                    onClick={() => flagMutation.mutate({ flag: f.name, enabled: !f.enabled, rollout_pct: f.rollout_pct })}
                    disabled={flagMutation.isPending}
                  >
                    {f.enabled ? "Disable" : "Enable"}
                  </Button>
                </div>
              </div>
            ))}
            {(flags?.data ?? []).length === 0 && <p className="text-gray-400 text-sm">No feature flags configured.</p>}
          </div>
        )}
      </div>

      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-2">Update Config (4-eyes)</h2>
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-800 mb-4">
          Config writes require a second authoriser with a different officer ID. All changes are versioned.
        </div>
        <div className="grid grid-cols-1 gap-3 max-w-lg">
          <Input label="Config Key" value={configKey} onChange={(e) => setConfigKey(e.target.value)} placeholder="e.g. kyc.tier1.daily_limit" />
          <Input label="Value (JSON or string)" value={configValue} onChange={(e) => setConfigValue(e.target.value)} placeholder='{"limit": 500000}' />
          <Input label="Second Authoriser ID" value={secondAuth} onChange={(e) => setSecondAuth(e.target.value)} placeholder="Officer UUID" />
          <Input label="Reason" value={reason} onChange={(e) => setReason(e.target.value)} placeholder="Reason for change" />
          {configError && <p className="text-red-500 text-sm">{configError}</p>}
          <Button onClick={() => configMutation.mutate()} disabled={configMutation.isPending}>
            {configMutation.isPending ? "Updating..." : "Update Config"}
          </Button>
        </div>
      </div>
    </div>
  );
}
