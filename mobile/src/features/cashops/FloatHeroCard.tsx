import { View, Text, StyleSheet, ActivityIndicator } from "react-native";
import { colors, spacing, radius } from "../../theme/tokens";

const PRIMARY = "#16A34A";

interface Props {
  floatMinor: number;
  limitMinor: number;
  agentStatus: string;
  isLoading: boolean;
}

const STATUS_COLOR: Record<string, string> = {
  ACTIVE:       "#16A34A",
  PENDING:      "#D97706",
  SUSPENDED:    "#DC2626",
  UNREGISTERED: "#6B7280",
};

export function FloatHeroCard({ floatMinor, limitMinor, agentStatus, isLoading }: Props) {
  const pct = limitMinor > 0 ? Math.min(Math.round((floatMinor / limitMinor) * 100), 100) : 0;
  const isLow = limitMinor > 0 && pct < 20;
  const floatETB = (floatMinor / 100).toFixed(2);
  const ceilingETB = limitMinor > 0 ? (limitMinor / 100).toFixed(0) : "—";
  const statusColor = STATUS_COLOR[agentStatus] ?? colors.textTertiary;

  return (
    <View style={styles.hero}>
      <Text style={styles.label}>FLOAT BALANCE</Text>
      {isLoading ? (
        <ActivityIndicator color="#FFF" style={{ marginVertical: spacing.lg }} />
      ) : (
        <>
          <Text style={styles.amount}>{floatETB} ETB</Text>
          <Text style={styles.ceiling}>Ceiling: {ceilingETB} ETB</Text>
          <View style={styles.barBg}>
            <View style={[styles.barFill, { width: `${pct}%` as any }, isLow && styles.barLow]} />
          </View>
          <View style={styles.footer}>
            <Text style={styles.pct}>{pct}% utilized</Text>
            <View style={[styles.badge, { backgroundColor: statusColor + "33" }]}>
              <Text style={[styles.badgeText, { color: statusColor }]}>{agentStatus}</Text>
            </View>
          </View>
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  hero: {
    backgroundColor: PRIMARY,
    borderRadius: radius.lg,
    padding: spacing.lg,
  },
  label: {
    fontSize: 11, fontWeight: "600",
    color: "rgba(255,255,255,0.7)",
    letterSpacing: 0.8, marginBottom: spacing.sm,
  },
  amount: { fontSize: 36, fontWeight: "800", color: "#FFF", marginBottom: 4 },
  ceiling: { fontSize: 13, color: "rgba(255,255,255,0.65)", marginBottom: spacing.md },
  barBg: { height: 8, backgroundColor: "rgba(255,255,255,0.25)", borderRadius: 4, marginBottom: 6 },
  barFill: { height: 8, backgroundColor: colors.gold, borderRadius: 4 },
  barLow: { backgroundColor: colors.error },
  footer: { flexDirection: "row", justifyContent: "space-between", alignItems: "center" },
  pct: { fontSize: 12, color: "rgba(255,255,255,0.6)" },
  badge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 3 },
  badgeText: { fontSize: 11, fontWeight: "700" },
});
