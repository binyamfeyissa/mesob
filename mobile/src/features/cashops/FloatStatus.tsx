import { View, Text, StyleSheet, ActivityIndicator } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow, typography } from "../../theme/tokens";

interface FloatData {
  agent_id: string;
  region_id: string;
  float_minor: number;
  float_limit_minor: number;
  status: string; // ACTIVE | PENDING | SUSPENDED | UNREGISTERED
}

interface FloatResponse {
  data: FloatData;
}

const AGENT_STATUS_COLOR: Record<string, string> = {
  ACTIVE: "#16A34A",
  PENDING: "#D97706",
  SUSPENDED: "#DC2626",
  UNREGISTERED: "#6B7280",
};

export function FloatStatus() {
  const { data, isLoading } = useQuery({
    queryKey: ["agent-float"],
    queryFn: (): Promise<FloatResponse> => apiFetch("/agent/float"),
    refetchInterval: 30_000,
  });

  const d = data?.data;
  const floatMinor = d?.float_minor ?? 0;
  const limitMinor = d?.float_limit_minor ?? 0;
  const agentStatus = d?.status ?? "UNREGISTERED";

  const pct = limitMinor > 0 ? Math.min(Math.round((floatMinor / limitMinor) * 100), 100) : 0;
  const isLow = limitMinor > 0 && pct < 20;

  const floatETB = (floatMinor / 100).toFixed(2);
  const ceilingETB = limitMinor > 0 ? (limitMinor / 100).toFixed(0) : "—";

  const statusColor = AGENT_STATUS_COLOR[agentStatus] ?? colors.textTertiary;

  return (
    <View style={styles.wrapper}>
      {/* Hero card */}
      <View style={[styles.heroCard, shadow.md]}>
        <Text style={styles.heroLabel}>FLOAT BALANCE</Text>
        {isLoading ? (
          <ActivityIndicator color="#FFF" style={{ marginVertical: spacing.lg }} />
        ) : (
          <>
            <Text style={styles.heroAmount}>{floatETB} ETB</Text>
            <Text style={styles.heroCeiling}>Ceiling: {ceilingETB} ETB</Text>

            <View style={styles.barBg}>
              <View
                style={[
                  styles.barFill,
                  { width: `${pct}%` as any },
                  isLow && styles.barLow,
                ]}
              />
            </View>
            <Text style={styles.barPct}>{pct}% utilized</Text>
          </>
        )}
      </View>

      {/* Low-float warning */}
      {isLow && (
        <View style={[styles.warningCard, shadow.sm]}>
          <Text style={styles.warningText}>Float is low — request a top-up from your branch.</Text>
        </View>
      )}

      {/* Info card */}
      <View style={[styles.infoCard, shadow.sm]}>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>Agent Status</Text>
          <View style={[styles.badge, { backgroundColor: statusColor + "22" }]}>
            <Text style={[styles.badgeText, { color: statusColor }]}>{agentStatus}</Text>
          </View>
        </View>
        <View style={[styles.infoRow, styles.infoRowBorder]}>
          <Text style={styles.infoLabel}>Float Health</Text>
          <View style={[styles.badge, { backgroundColor: isLow ? colors.errorLight : colors.successLight }]}>
            <Text style={[styles.badgeText, { color: isLow ? colors.error : colors.success }]}>
              {isLow ? "LOW" : "HEALTHY"}
            </Text>
          </View>
        </View>
        <View style={[styles.infoRow, styles.infoRowBorder]}>
          <Text style={styles.infoLabel}>Available</Text>
          <Text style={styles.infoValue}>{floatETB} ETB</Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>Ceiling</Text>
          <Text style={styles.infoValue}>{ceilingETB} ETB</Text>
        </View>
      </View>
    </View>
  );
}

const AGENT_PRIMARY = "#16A34A";

const styles = StyleSheet.create({
  wrapper: { paddingHorizontal: spacing.md, gap: spacing.md },

  heroCard: {
    backgroundColor: AGENT_PRIMARY,
    borderRadius: radius.lg,
    padding: spacing.lg,
  },
  heroLabel: { fontSize: 11, fontWeight: "600", color: "rgba(255,255,255,0.7)", letterSpacing: 0.8, marginBottom: spacing.sm },
  heroAmount: { fontSize: 36, fontWeight: "800", color: "#FFF", marginBottom: 4 },
  heroCeiling: { fontSize: 13, color: "rgba(255,255,255,0.65)", marginBottom: spacing.md },

  barBg: { height: 8, backgroundColor: "rgba(255,255,255,0.25)", borderRadius: 4, marginBottom: 6 },
  barFill: { height: 8, backgroundColor: colors.gold, borderRadius: 4 },
  barLow: { backgroundColor: colors.error },
  barPct: { fontSize: 12, color: "rgba(255,255,255,0.6)" },

  warningCard: {
    backgroundColor: colors.errorLight,
    borderRadius: radius.md,
    padding: spacing.md,
  },
  warningText: { fontSize: 13, color: colors.error, fontWeight: "500" },

  infoCard: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },
  infoRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", padding: spacing.md },
  infoRowBorder: { borderTopWidth: 1, borderColor: colors.divider },
  infoLabel: { fontSize: 14, color: colors.textSecondary },
  infoValue: { fontSize: 14, fontWeight: "600", color: colors.text },
  badge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 3 },
  badgeText: { fontSize: 11, fontWeight: "700" },
});
