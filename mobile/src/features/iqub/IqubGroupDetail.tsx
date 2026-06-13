import { useState } from "react";
import {
  View, Text, TouchableOpacity, StyleSheet, ScrollView,
  Alert, ActivityIndicator, Clipboard,
} from "react-native";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

export interface IqubGroupData {
  group_id: string;
  name: string;
  cycle_minor: number;
  frequency: string;
  member_limit: number;
  cycle?: {
    id: string;
    number: number;
    paid: number;
    total: number;
    next_payout_member: string;
    due_date: string;
  };
}

interface MemberRow {
  membership_id: string;
  user_id: string;
  payout_order: number;
  cycle_state: string;
}

interface Props {
  group: IqubGroupData;
  currentUserId: string;
  onBack: () => void;
}

const STATE_COLOR: Record<string, string> = {
  PAID: colors.success,
  PENDING: colors.textTertiary,
  MISSED: colors.error,
  RECEIVED: colors.primary,
};

export function IqubGroupDetail({ group, currentUserId, onBack }: Props) {
  const [activeTab, setActiveTab] = useState<"overview" | "members">("overview");
  const [contributing, setContributing] = useState(false);
  const [closingCycle, setClosingCycle] = useState(false);
  const [joinCode, setJoinCode] = useState("");
  const queryClient = useQueryClient();

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ["iqub-members", group.group_id],
    queryFn: () => apiFetch<{ data: MemberRow[] }>(`/iqub/groups/${group.group_id}/members`),
    enabled: activeTab === "members",
  });
  const members = membersData?.data ?? [];

  // ── Contribute ──────────────────────────────────────────────────────────────
  async function handleContribute() {
    if (!group.cycle?.id) {
      Alert.alert("No Active Cycle", "There is no open cycle to contribute to.");
      return;
    }
    setContributing(true);
    try {
      await apiFetch(`/iqub/groups/${group.group_id}/contribute`, {
        method: "POST",
        body: JSON.stringify({
          cycle_id: group.cycle.id,
          idempotency_key: generateCaptureKey(),
        }),
      });
      queryClient.invalidateQueries({ queryKey: ["iqub-groups"] });
      Alert.alert("Contributed!", `Cycle ${group.cycle.number} contribution recorded.`);
    } catch (e: any) {
      Alert.alert("Failed", e.message);
    } finally {
      setContributing(false);
    }
  }

  // ── Request Payout (close cycle) ────────────────────────────────────────────
  async function handleRequestPayout() {
    if (!group.cycle?.id) {
      Alert.alert("No Active Cycle", "There is no open cycle to close.");
      return;
    }
    Alert.alert(
      "Close Cycle & Disburse Payout",
      `This will close cycle ${group.cycle.number} and disburse the payout to the recipient. Continue?`,
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Close Cycle",
          style: "destructive",
          onPress: async () => {
            setClosingCycle(true);
            try {
              await apiFetch(`/iqub/cycles/${group.cycle!.id}/close`, { method: "POST" });
              queryClient.invalidateQueries({ queryKey: ["iqub-groups"] });
              Alert.alert("Cycle Closed", "The payout has been disbursed.");
              onBack();
            } catch (e: any) {
              Alert.alert("Failed", e.message);
            } finally {
              setClosingCycle(false);
            }
          },
        },
      ]
    );
  }

  // ── Copy join code ───────────────────────────────────────────────────────────
  async function handleFetchAndCopyJoinCode() {
    try {
      const res = await apiFetch<{ data: { join_code: string } }>(`/iqub/groups/${group.group_id}`);
      const code = res.data?.join_code ?? "";
      setJoinCode(code);
      Clipboard.setString(code);
      Alert.alert("Copied!", `Join code "${code}" copied to clipboard.\nShare it with new members along with the Group ID.`);
    } catch (e: any) {
      Alert.alert("Error", e.message);
    }
  }

  const hasCycle = !!group.cycle?.id;
  const paidFraction = hasCycle && group.cycle!.total > 0
    ? group.cycle!.paid / group.cycle!.total
    : 0;

  return (
    <ScrollView
      contentContainerStyle={styles.container}
      showsVerticalScrollIndicator={false}
      keyboardShouldPersistTaps="handled"
    >
      {/* Hero card */}
      <View style={[styles.hero, shadow.sm]}>
        <View style={styles.heroTop}>
          <Text style={styles.heroName}>{group.name}</Text>
          <View style={styles.freqBadge}>
            <Text style={styles.freqText}>{group.frequency}</Text>
          </View>
        </View>
        <Text style={styles.heroAmount}>
          {(group.cycle_minor / 100).toFixed(0)} ETB / cycle
        </Text>
        <Text style={styles.heroSub}>Up to {group.member_limit} members</Text>
      </View>

      {/* Cycle progress */}
      {hasCycle ? (
        <View style={[styles.card, shadow.sm]}>
          <View style={styles.cycleHeader}>
            <Text style={styles.cardTitle}>Cycle {group.cycle!.number}</Text>
            <Text style={styles.cycleDue}>Due {group.cycle!.due_date}</Text>
          </View>
          <View style={styles.progressBar}>
            <View style={[styles.progressFill, { width: `${Math.round(paidFraction * 100)}%` as any }]} />
          </View>
          <Text style={styles.cycleStats}>
            {group.cycle!.paid} of {group.cycle!.total} members paid
          </Text>
        </View>
      ) : (
        <View style={[styles.card, shadow.sm, styles.noCycleCard]}>
          <Text style={styles.noCycleText}>No active cycle — group is forming</Text>
        </View>
      )}

      {/* Actions */}
      <View style={styles.actionsRow}>
        <TouchableOpacity
          style={[styles.actionBtn, styles.contributeBtn, (!hasCycle || contributing) && styles.btnDisabled]}
          onPress={handleContribute}
          disabled={!hasCycle || contributing}
          accessibilityRole="button"
        >
          {contributing
            ? <ActivityIndicator color="#FFF" size="small" />
            : <Text style={styles.actionBtnText}>Contribute</Text>
          }
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.actionBtn, styles.payoutBtn, (!hasCycle || closingCycle) && styles.btnDisabled]}
          onPress={handleRequestPayout}
          disabled={!hasCycle || closingCycle}
          accessibilityRole="button"
        >
          {closingCycle
            ? <ActivityIndicator color="#FFF" size="small" />
            : <Text style={styles.actionBtnText}>Request Payout</Text>
          }
        </TouchableOpacity>
      </View>

      {/* Invite / join code */}
      <View style={[styles.card, shadow.sm]}>
        <Text style={styles.cardTitle}>Invite Members</Text>
        <Text style={styles.inviteHint}>
          Share the join code and Group ID with people you want to invite.
        </Text>
        {joinCode ? (
          <View style={styles.codeBox}>
            <Text style={styles.codeText}>{joinCode}</Text>
          </View>
        ) : null}
        <TouchableOpacity
          style={[styles.actionBtn, styles.inviteBtn]}
          onPress={handleFetchAndCopyJoinCode}
          accessibilityRole="button"
        >
          <Text style={styles.actionBtnText}>Copy Join Code</Text>
        </TouchableOpacity>
        <View style={styles.groupIdRow}>
          <Text style={styles.groupIdLabel}>Group ID</Text>
          <Text style={styles.groupIdValue} selectable>{group.group_id}</Text>
        </View>
      </View>

      {/* Members tab */}
      <View style={[styles.card, shadow.sm]}>
        <Text style={styles.cardTitle}>Members</Text>
        {membersLoading && <ActivityIndicator color={colors.primary} style={{ margin: 16 }} />}
        {!membersLoading && members.length === 0 && (
          <Text style={styles.emptyText}>No members yet.</Text>
        )}
        {members.map((m) => (
          <View key={m.membership_id} style={styles.memberRow}>
            <Text style={styles.memberId} numberOfLines={1}>
              {m.user_id === currentUserId ? "You" : `…${m.user_id.slice(-6)}`}
            </Text>
            <View style={[styles.stateBadge, { backgroundColor: (STATE_COLOR[m.cycle_state] ?? colors.textTertiary) + "22" }]}>
              <Text style={[styles.stateText, { color: STATE_COLOR[m.cycle_state] ?? colors.textTertiary }]}>
                {m.cycle_state}
              </Text>
            </View>
          </View>
        ))}
        {!membersLoading && members.length === 0 && (
          <TouchableOpacity
            style={[styles.actionBtn, styles.inviteBtn, { marginTop: spacing.sm }]}
            onPress={handleFetchAndCopyJoinCode}
            accessibilityRole="button"
          >
            <Text style={styles.actionBtnText}>Invite First Member</Text>
          </TouchableOpacity>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: spacing.md, gap: spacing.md, paddingBottom: 32 },

  hero: {
    backgroundColor: colors.primary,
    borderRadius: radius.lg,
    padding: spacing.lg,
  },
  heroTop: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", marginBottom: 6 },
  heroName: { fontSize: 20, fontWeight: "700", color: "#FFF", flex: 1 },
  freqBadge: { backgroundColor: "rgba(255,255,255,0.2)", borderRadius: 8, paddingHorizontal: 10, paddingVertical: 3 },
  freqText: { fontSize: 12, fontWeight: "600", color: "#FFF" },
  heroAmount: { fontSize: 28, fontWeight: "800", color: "#FFF", marginBottom: 2 },
  heroSub: { fontSize: 13, color: "rgba(255,255,255,0.75)" },

  card: { backgroundColor: colors.surface, borderRadius: radius.lg, padding: spacing.md },
  cardTitle: { fontSize: 14, fontWeight: "700", color: colors.text, marginBottom: spacing.sm },

  cycleHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  cycleDue: { fontSize: 12, color: colors.textTertiary },
  progressBar: { height: 8, backgroundColor: colors.bg, borderRadius: 4, overflow: "hidden", marginBottom: 6 },
  progressFill: { height: "100%", backgroundColor: colors.success, borderRadius: 4 },
  cycleStats: { fontSize: 13, color: colors.textSecondary },

  noCycleCard: { alignItems: "center", paddingVertical: spacing.lg },
  noCycleText: { fontSize: 13, color: colors.textTertiary },

  actionsRow: { flexDirection: "row", gap: spacing.sm },
  actionBtn: {
    flex: 1, borderRadius: radius.md, height: 48,
    alignItems: "center", justifyContent: "center",
  },
  contributeBtn: { backgroundColor: colors.success },
  payoutBtn: { backgroundColor: colors.primary },
  inviteBtn: { backgroundColor: colors.navy, marginTop: spacing.sm },
  btnDisabled: { opacity: 0.45 },
  actionBtnText: { fontSize: 14, fontWeight: "700", color: "#FFF" },

  inviteHint: { fontSize: 13, color: colors.textSecondary, marginBottom: spacing.sm, lineHeight: 18 },
  codeBox: {
    backgroundColor: colors.bg, borderRadius: radius.sm,
    paddingVertical: 12, alignItems: "center", marginBottom: spacing.sm,
  },
  codeText: { fontSize: 24, fontWeight: "800", letterSpacing: 6, color: colors.navy },

  groupIdRow: { marginTop: spacing.sm, flexDirection: "row", alignItems: "center", gap: spacing.xs },
  groupIdLabel: { fontSize: 11, fontWeight: "600", color: colors.textTertiary },
  groupIdValue: { fontSize: 11, color: colors.textTertiary, flex: 1 },

  memberRow: {
    flexDirection: "row", alignItems: "center", justifyContent: "space-between",
    paddingVertical: 10, borderTopWidth: 1, borderTopColor: colors.divider,
  },
  memberId: { fontSize: 13, color: colors.text, flex: 1 },
  stateBadge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 3 },
  stateText: { fontSize: 11, fontWeight: "600" },
  emptyText: { fontSize: 13, color: colors.textTertiary, paddingVertical: spacing.sm },
});
