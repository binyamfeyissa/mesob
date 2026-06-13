import { useState } from "react";
import {
  View, Text, TouchableOpacity, StyleSheet, ScrollView,
  Alert, ActivityIndicator,
} from "react-native";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { FileClaimForm } from "./FileClaimForm";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

export interface IddirGroupData {
  group_id: string;
  name: string;
  premium_minor: number;
  benefit_minor: number;
  frequency: string;
  coverage_status: string;
}

interface ClaimRow {
  id: string;
  type: string;
  status: string;
  settled_minor: number;
  created_at: string;
}

const CLAIM_COLOR: Record<string, string> = {
  UNDER_REVIEW: colors.warning,
  APPROVED: colors.success,
  REJECTED: colors.error,
  SETTLED: colors.primary,
};

function currentPeriod(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

interface Props {
  group: IddirGroupData;
  onBack: () => void;
}

export function IddirGroupDetail({ group, onBack }: Props) {
  const [filingClaim, setFilingClaim] = useState(false);
  const [payingPremium, setPayingPremium] = useState(false);
  const queryClient = useQueryClient();

  const { data: claimsData, isLoading: claimsLoading, refetch: refetchClaims } = useQuery({
    queryKey: ["iddir-claims", group.group_id],
    queryFn: () => apiFetch<{ data: ClaimRow[] }>(`/iddir/groups/${group.group_id}/claims`),
  });
  const claims = claimsData?.data ?? [];

  // ── Pay Premium ─────────────────────────────────────────────────────────────
  async function handlePayPremium() {
    const period = currentPeriod();
    Alert.alert(
      "Pay Premium",
      `Pay ${(group.premium_minor / 100).toFixed(2)} ETB for period ${period}?`,
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Pay",
          onPress: async () => {
            setPayingPremium(true);
            try {
              await apiFetch(`/iddir/groups/${group.group_id}/premium`, {
                method: "POST",
                body: JSON.stringify({ period, idempotency_key: generateCaptureKey() }),
              });
              queryClient.invalidateQueries({ queryKey: ["iddir-groups"] });
              Alert.alert("Premium Paid", `Coverage renewed for ${period}.`);
            } catch (e: any) {
              Alert.alert("Failed", e.message);
            } finally {
              setPayingPremium(false);
            }
          },
        },
      ]
    );
  }

  const isActive = group.coverage_status === "ACTIVE";

  if (filingClaim) {
    return (
      <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
        <View style={[styles.card, shadow.sm]}>
          <FileClaimForm
            groupId={group.group_id}
            onFiled={(claimId) => {
              setFilingClaim(false);
              refetchClaims();
              Alert.alert("Claim Filed", `Claim ID: …${claimId.slice(-8)}\nWe will review and notify you.`);
            }}
          />
          <TouchableOpacity style={styles.cancelLink} onPress={() => setFilingClaim(false)}>
            <Text style={styles.cancelText}>Cancel</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    );
  }

  return (
    <ScrollView
      contentContainerStyle={styles.container}
      showsVerticalScrollIndicator={false}
    >
      {/* Hero */}
      <View style={[styles.hero, shadow.sm]}>
        <View style={styles.heroTop}>
          <Text style={styles.heroName}>{group.name}</Text>
          <View style={[styles.coverBadge, isActive ? styles.coverActive : styles.coverInactive]}>
            <Text style={styles.coverText}>{group.coverage_status}</Text>
          </View>
        </View>
        <Text style={styles.heroAmount}>
          {(group.benefit_minor / 100).toFixed(0)} ETB benefit
        </Text>
        <Text style={styles.heroSub}>
          Premium: {(group.premium_minor / 100).toFixed(0)} ETB · {group.frequency}
        </Text>
      </View>

      {/* Actions */}
      <View style={styles.actionsRow}>
        <TouchableOpacity
          style={[styles.actionBtn, styles.premiumBtn, payingPremium && styles.btnDisabled]}
          onPress={handlePayPremium}
          disabled={payingPremium}
          accessibilityRole="button"
        >
          {payingPremium
            ? <ActivityIndicator color="#FFF" size="small" />
            : <Text style={styles.actionBtnText}>Pay Premium</Text>
          }
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.actionBtn, styles.claimBtn]}
          onPress={() => setFilingClaim(true)}
          accessibilityRole="button"
        >
          <Text style={styles.actionBtnText}>File a Claim</Text>
        </TouchableOpacity>
      </View>

      {/* Claims history */}
      <View style={[styles.card, shadow.sm]}>
        <Text style={styles.cardTitle}>My Claims</Text>
        {claimsLoading && <ActivityIndicator color={colors.primary} style={{ margin: 16 }} />}
        {!claimsLoading && claims.length === 0 && (
          <Text style={styles.emptyText}>No claims filed yet.</Text>
        )}
        {claims.map((c) => {
          const color = CLAIM_COLOR[c.status] ?? colors.textTertiary;
          return (
            <View key={c.id} style={styles.claimRow}>
              <View style={styles.claimLeft}>
                <Text style={styles.claimType}>{c.type}</Text>
                <Text style={styles.claimDate}>
                  {new Date(c.created_at).toLocaleDateString()}
                </Text>
              </View>
              <View style={styles.claimRight}>
                <View style={[styles.statusBadge, { backgroundColor: color + "22" }]}>
                  <Text style={[styles.statusText, { color }]}>{c.status}</Text>
                </View>
                {(c.settled_minor ?? 0) > 0 && (
                  <Text style={styles.settled}>
                    {(c.settled_minor / 100).toFixed(0)} ETB
                  </Text>
                )}
              </View>
            </View>
          );
        })}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: spacing.md, gap: spacing.md, paddingBottom: 32 },

  hero: {
    backgroundColor: colors.success,
    borderRadius: radius.lg,
    padding: spacing.lg,
  },
  heroTop: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", marginBottom: 6 },
  heroName: { fontSize: 20, fontWeight: "700", color: "#FFF", flex: 1 },
  coverBadge: { borderRadius: 8, paddingHorizontal: 10, paddingVertical: 3 },
  coverActive: { backgroundColor: "rgba(255,255,255,0.25)" },
  coverInactive: { backgroundColor: "rgba(220,38,38,0.5)" },
  coverText: { fontSize: 12, fontWeight: "600", color: "#FFF" },
  heroAmount: { fontSize: 28, fontWeight: "800", color: "#FFF", marginBottom: 2 },
  heroSub: { fontSize: 13, color: "rgba(255,255,255,0.8)" },

  actionsRow: { flexDirection: "row", gap: spacing.sm },
  actionBtn: {
    flex: 1, borderRadius: radius.md, height: 48,
    alignItems: "center", justifyContent: "center",
  },
  premiumBtn: { backgroundColor: colors.success },
  claimBtn: { backgroundColor: colors.error },
  btnDisabled: { opacity: 0.45 },
  actionBtnText: { fontSize: 14, fontWeight: "700", color: "#FFF" },

  card: { backgroundColor: colors.surface, borderRadius: radius.lg, padding: spacing.md },
  cardTitle: { fontSize: 14, fontWeight: "700", color: colors.text, marginBottom: spacing.sm },

  claimRow: {
    flexDirection: "row", alignItems: "center", justifyContent: "space-between",
    paddingVertical: 12, borderTopWidth: 1, borderTopColor: colors.divider,
  },
  claimLeft: { flex: 1 },
  claimType: { fontSize: 14, fontWeight: "600", color: colors.text },
  claimDate: { fontSize: 12, color: colors.textTertiary, marginTop: 2 },
  claimRight: { alignItems: "flex-end", gap: 4 },
  statusBadge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 3 },
  statusText: { fontSize: 11, fontWeight: "600" },
  settled: { fontSize: 13, fontWeight: "700", color: colors.success },

  emptyText: { fontSize: 13, color: colors.textTertiary, paddingVertical: spacing.sm },
  cancelLink: { paddingTop: spacing.md, alignItems: "center" },
  cancelText: { fontSize: 14, color: colors.textSecondary },
});
