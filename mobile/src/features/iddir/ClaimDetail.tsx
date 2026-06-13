import { View, Text, StyleSheet, ActivityIndicator } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";

interface ClaimDetailProps {
  groupId: string;
  claimId: string;
}

interface ClaimResponse {
  data: {
    claim_id: string;
    status: string;
    settled_minor: number;
    transaction_id: string | null;
  };
}

const statusColor = (s: string) =>
  ({ UNDER_REVIEW: "#F59E0B", APPROVED: "#059669", REJECTED: "#DC2626", SETTLED: "#1B4FDE" }[s] ?? "#6B7280");

export function ClaimDetail({ groupId, claimId }: ClaimDetailProps) {
  const { data, isLoading } = useQuery({
    queryKey: ["claim", groupId, claimId],
    queryFn: () => apiFetch<ClaimResponse>(`/iddir/groups/${groupId}/claims/${claimId}`),
    refetchInterval: 30_000,
  });

  if (isLoading) return <ActivityIndicator style={{ margin: 32 }} />;

  const c = data?.data;
  return (
    <View style={styles.card}>
      <Text style={styles.id}>Claim {c?.claim_id?.slice(0, 8)}...</Text>
      <View style={[styles.badge, { backgroundColor: statusColor(c?.status ?? "") + "22" }]}>
        <Text style={[styles.badgeText, { color: statusColor(c?.status ?? "") }]}>{c?.status}</Text>
      </View>
      {(c?.settled_minor ?? 0) > 0 && (
        <Text style={styles.settled}>Settlement: {((c!.settled_minor) / 100).toFixed(2)} ETB</Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  card: { margin: 16, backgroundColor: "#FFF", borderRadius: 16, padding: 20 },
  id: { fontSize: 14, color: "#6B7280", marginBottom: 12 },
  badge: { alignSelf: "flex-start", borderRadius: 8, paddingHorizontal: 12, paddingVertical: 6, marginBottom: 12 },
  badgeText: { fontWeight: "700", fontSize: 14 },
  settled: { fontSize: 16, fontWeight: "700", color: "#059669" },
});
