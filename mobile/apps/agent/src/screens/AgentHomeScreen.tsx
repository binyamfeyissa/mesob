import { ScrollView, View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter } from "expo-router";
import { HugeiconsIcon } from "@hugeicons/react-native";
import type { IconSvgElement } from "@hugeicons/react-native";
import {
  ArrowDown01Icon,
  ArrowUp01Icon,
  UserAdd01Icon,
  ArrowReloadHorizontalIcon,
  Notification01Icon,
} from "@hugeicons/core-free-icons";
import { useQuery } from "@tanstack/react-query";
import { FloatHeroCard } from "../../../../src/features/cashops/FloatHeroCard";
import { SyncStatus } from "../../../../src/features/cashops/SyncStatus";
import { useAuthStore } from "../../../../src/lib/authStore";
import { apiFetch } from "../../../../src/api/client";
import { operationQueue } from "../../../../src/sync/queue";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";

const PRIMARY = "#16A34A";

interface FloatData {
  agent_id: string;
  float_minor: number;
  float_limit_minor: number;
  status: string;
}

function initials(msisdn: string) {
  return msisdn ? msisdn.slice(-2) : "AG";
}

function QuickAction({
  icon,
  label,
  color,
  onPress,
  badge,
}: {
  icon: IconSvgElement;
  label: string;
  color: string;
  onPress: () => void;
  badge?: number;
}) {
  return (
    <TouchableOpacity style={styles.actionBtn} onPress={onPress} accessibilityRole="button">
      <View style={[styles.actionIcon, { backgroundColor: color + "18" }]}>
        <HugeiconsIcon icon={icon} size={22} color={color} strokeWidth={1.8} />
        {badge != null && badge > 0 && (
          <View style={styles.badge}>
            <Text style={styles.badgeText}>{badge > 9 ? "9+" : badge}</Text>
          </View>
        )}
      </View>
      <Text style={styles.actionLabel}>{label}</Text>
    </TouchableOpacity>
  );
}

export function AgentHomeScreen() {
  const { user } = useAuthStore();
  const router = useRouter();
  const pendingCount = operationQueue.getPending().length;

  const { data } = useQuery({
    queryKey: ["agent-float"],
    queryFn: (): Promise<{ data: FloatData }> => apiFetch("/agent/float"),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  const d = data?.data;
  const floatMinor = d?.float_minor ?? 0;
  const limitMinor = d?.float_limit_minor ?? 0;
  const agentStatus = d?.status ?? "UNREGISTERED";

  const isLow = limitMinor > 0 && floatMinor / limitMinor < 0.2;

  const actions: { icon: IconSvgElement; label: string; color: string; route: string; badge?: number }[] = [
    { icon: ArrowDown01Icon,              label: "Cash In",  color: PRIMARY,          route: "/(tabs)/operations" },
    { icon: ArrowUp01Icon,                label: "Cash Out", color: colors.warning,    route: "/(tabs)/operations" },
    { icon: UserAdd01Icon,                label: "Onboard",  color: colors.primary,    route: "/(tabs)/onboard" },
    { icon: ArrowReloadHorizontalIcon,    label: "Sync",     color: colors.textSecondary, route: "/(tabs)/sync-progress", badge: pendingCount },
  ];

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.avatarCircle}>
            <Text style={styles.avatarText}>{initials(user?.msisdn ?? "")}</Text>
          </View>
          <View style={styles.headerCenter}>
            <Text style={styles.greeting}>Agent Dashboard</Text>
            <Text style={styles.msisdn}>{user?.msisdn ?? ""}</Text>
          </View>
          <TouchableOpacity style={styles.iconBtn} accessibilityRole="button" accessibilityLabel="Notifications">
            <HugeiconsIcon icon={Notification01Icon} size={20} color={colors.text} strokeWidth={1.8} />
          </TouchableOpacity>
        </View>

        {/* Float hero card */}
        <View style={[styles.card, shadow.sm, { marginBottom: spacing.lg }]}>
          <FloatHeroCard
            floatMinor={floatMinor}
            limitMinor={limitMinor}
            agentStatus={agentStatus}
            isLoading={!data}
          />
        </View>

        {/* Low-float warning */}
        {isLow && (
          <View style={[styles.warningBanner, shadow.sm]}>
            <Text style={styles.warningText}>Float is low — request a top-up from your branch.</Text>
          </View>
        )}

        {/* Quick actions */}
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionLabel}>QUICK ACTIONS</Text>
        </View>
        <View style={[styles.card, shadow.sm, styles.actionsCard]}>
          {actions.map((a) => (
            <QuickAction
              key={a.label}
              icon={a.icon}
              label={a.label}
              color={a.color}
              badge={a.badge}
              onPress={() => router.push(a.route as any)}
            />
          ))}
        </View>

        {/* Offline queue */}
        <View style={[styles.sectionHeader, { marginTop: spacing.lg }]}>
          <Text style={styles.sectionLabel}>OFFLINE QUEUE</Text>
        </View>
        <SyncStatus />

        <View style={{ height: 120 }} />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
  scroll: { flex: 1 },
  content: { paddingTop: spacing.md, paddingHorizontal: spacing.md },

  header: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: spacing.lg,
    gap: spacing.sm,
  },
  avatarCircle: {
    width: 42, height: 42, borderRadius: 21,
    backgroundColor: PRIMARY + "22",
    borderWidth: 1, borderColor: PRIMARY + "44",
    alignItems: "center", justifyContent: "center",
  },
  avatarText: { fontSize: 14, fontWeight: "700", color: PRIMARY },
  headerCenter: { flex: 1 },
  greeting: { fontSize: 13, fontWeight: "600", color: colors.text },
  msisdn: { fontSize: 12, color: colors.textTertiary, marginTop: 1 },
  iconBtn: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.surface, alignItems: "center", justifyContent: "center", ...shadow.sm,
  },

  card: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },
  actionsCard: {
    flexDirection: "row", paddingVertical: spacing.lg,
    paddingHorizontal: spacing.md, justifyContent: "space-around",
  },

  actionBtn: { alignItems: "center", gap: spacing.sm },
  actionIcon: {
    width: 56, height: 56, borderRadius: 28,
    alignItems: "center", justifyContent: "center",
  },
  actionLabel: { fontSize: 12, fontWeight: "500", color: colors.textSecondary },

  badge: {
    position: "absolute", top: -2, right: -2,
    backgroundColor: colors.error, borderRadius: 8,
    minWidth: 16, height: 16, paddingHorizontal: 3,
    alignItems: "center", justifyContent: "center",
  },
  badgeText: { fontSize: 9, fontWeight: "700", color: "#FFF" },

  warningBanner: {
    backgroundColor: colors.errorLight,
    borderRadius: radius.md,
    padding: spacing.md,
    marginBottom: spacing.md,
  },
  warningText: { fontSize: 13, color: colors.error, fontWeight: "500" },

  sectionHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: spacing.sm },
  sectionLabel: { ...typography.label },
});
