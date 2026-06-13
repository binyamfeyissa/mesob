import { View, Text, TouchableOpacity, StyleSheet, ScrollView, Alert, ActivityIndicator } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { Logout01Icon, UserCircleIcon, Location01Icon, Wallet01Icon, ArrowReloadHorizontalIcon } from "@hugeicons/core-free-icons";
import { useAuthStore } from "../../../../src/lib/authStore";
import { apiFetch, getAccessToken } from "../../../../src/api/client";
import { operationQueue } from "../../../../src/sync/queue";
import { syncPendingOperations } from "../../../../src/sync/engine";
import { colors, spacing, radius, shadow } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AgentNavigator";

const AGENT_PRIMARY = "#16A34A";
const AGENT_PRIMARY_LIGHT = "rgba(22,163,74,0.12)";

interface FloatData {
  agent_id: string;
  region_id: string;
  float_minor: number;
  float_limit_minor: number;
  status: string;
}

const AGENT_STATUS_COLOR: Record<string, string> = {
  ACTIVE: "#16A34A",
  PENDING: "#D97706",
  SUSPENDED: "#DC2626",
  UNREGISTERED: "#6B7280",
};

export function ProfileScreen() {
  const { user, logout } = useAuthStore();
  const [syncing, setSyncing] = useState(false);
  const [pendingCount, setPendingCount] = useState(() => operationQueue.getPending().length);

  const handleSync = async () => {
    const token = getAccessToken();
    if (!token) { Alert.alert("Not signed in"); return; }
    setSyncing(true);
    try {
      await syncPendingOperations(token);
      setPendingCount(operationQueue.getPending().length);
      Alert.alert("Sync complete");
    } catch {
      Alert.alert("Sync failed", "Check your connection and try again.");
    } finally {
      setSyncing(false);
    }
  };

  const { data } = useQuery({
    queryKey: ["agent-float"],
    queryFn: (): Promise<{ data: FloatData }> => apiFetch("/agent/float"),
    staleTime: 60_000,
  });

  const agentData = data?.data;
  const agentStatus = agentData?.status ?? "UNREGISTERED";
  const statusColor = AGENT_STATUS_COLOR[agentStatus] ?? colors.textTertiary;

  const handleLogout = () => {
    Alert.alert(
      "Log Out",
      "Are you sure you want to log out?",
      [
        { text: "Cancel", style: "cancel" },
        { text: "Log Out", style: "destructive", onPress: logout },
      ]
    );
  };

  const shortId = agentData?.agent_id
    ? agentData.agent_id.slice(0, 8) + "…"
    : "—";

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        contentContainerStyle={{ paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }}
        showsVerticalScrollIndicator={false}
      >
        {/* Avatar + name card */}
        <View style={[styles.heroCard, shadow.md]}>
          <View style={styles.avatarCircle}>
            <HugeiconsIcon icon={UserCircleIcon} size={44} color="#FFF" strokeWidth={1.6} />
          </View>
          <Text style={styles.heroMsisdn}>{user?.msisdn ?? "—"}</Text>
          <View style={[styles.statusBadge, { backgroundColor: statusColor + "33" }]}>
            <Text style={[styles.statusBadgeText, { color: statusColor }]}>{agentStatus}</Text>
          </View>
        </View>

        {/* Details card */}
        <View style={[styles.card, shadow.sm]}>
          <Row icon={UserCircleIcon} label="Agent ID" value={shortId} />
          <Divider />
          <Row icon={Location01Icon} label="Region" value={agentData?.region_id ? agentData.region_id.slice(0, 8) + "…" : "—"} />
          <Divider />
          <Row icon={Wallet01Icon} label="Role" value={user?.role ?? "AGENT"} />
        </View>

        {/* Sync */}
        <TouchableOpacity
          style={[styles.syncBtn, shadow.sm, syncing && styles.btnDisabled]}
          onPress={handleSync}
          disabled={syncing}
          accessibilityRole="button"
        >
          {syncing ? (
            <ActivityIndicator color={AGENT_PRIMARY} size="small" />
          ) : (
            <HugeiconsIcon icon={ArrowReloadHorizontalIcon} size={18} color={AGENT_PRIMARY} strokeWidth={2} />
          )}
          <Text style={styles.syncText}>
            Sync Offline Queue {pendingCount > 0 ? `(${pendingCount} pending)` : ""}
          </Text>
        </TouchableOpacity>

        {/* Logout */}
        <TouchableOpacity style={[styles.logoutBtn, shadow.sm]} onPress={handleLogout} accessibilityRole="button">
          <HugeiconsIcon icon={Logout01Icon} size={18} color={colors.error} strokeWidth={2} />
          <Text style={styles.logoutText}>Log Out</Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}

function Row({
  icon,
  label,
  value,
}: {
  icon: any;
  label: string;
  value: string;
}) {
  return (
    <View style={styles.row}>
      <View style={styles.rowIcon}>
        <HugeiconsIcon icon={icon} size={16} color={AGENT_PRIMARY} strokeWidth={1.8} />
      </View>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue} numberOfLines={1}>{value}</Text>
    </View>
  );
}

function Divider() {
  return <View style={styles.divider} />;
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },

  heroCard: {
    margin: spacing.md,
    backgroundColor: AGENT_PRIMARY,
    borderRadius: radius.lg,
    padding: spacing.lg,
    alignItems: "center",
  },
  avatarCircle: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: "rgba(255,255,255,0.2)",
    alignItems: "center",
    justifyContent: "center",
    marginBottom: spacing.md,
  },
  heroMsisdn: { fontSize: 20, fontWeight: "700", color: "#FFF", marginBottom: spacing.sm },
  statusBadge: { borderRadius: 12, paddingHorizontal: 12, paddingVertical: 4 },
  statusBadgeText: { fontSize: 12, fontWeight: "700" },

  card: {
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    overflow: "hidden",
  },

  row: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 14,
    paddingHorizontal: spacing.md,
  },
  rowIcon: {
    width: 32,
    height: 32,
    borderRadius: 8,
    backgroundColor: AGENT_PRIMARY_LIGHT,
    alignItems: "center",
    justifyContent: "center",
    marginRight: spacing.md,
  },
  rowLabel: { flex: 1, fontSize: 14, color: colors.textSecondary },
  rowValue: { fontSize: 14, fontWeight: "600", color: colors.text, maxWidth: 180 },

  divider: { height: 1, backgroundColor: colors.divider, marginLeft: 32 + spacing.md + spacing.md },

  syncBtn: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    backgroundColor: AGENT_PRIMARY_LIGHT,
    borderRadius: radius.md,
    height: 52,
    gap: spacing.sm,
  },
  syncText: { fontSize: 15, fontWeight: "600", color: AGENT_PRIMARY },
  btnDisabled: { opacity: 0.55 },
  logoutBtn: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    marginHorizontal: spacing.md,
    backgroundColor: colors.errorLight,
    borderRadius: radius.md,
    height: 52,
    gap: spacing.sm,
  },
  logoutText: { fontSize: 15, fontWeight: "700", color: colors.error },
});
