import { View, Text, TouchableOpacity, StyleSheet, ActivityIndicator } from "react-native";
import type { IddirGroupData } from "./IddirGroupDetail";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

interface IddirGroup {
  group_id: string;
  name: string;
  premium_minor: number;
  benefit_minor: number;
  frequency: string;
  coverage_status: string;
}

async function fetchGroups(): Promise<{ data: IddirGroup[] }> {
  return apiFetch("/iddir/groups");
}

function currentPeriod(): string {
  const now = new Date();
  const mm = String(now.getMonth() + 1).padStart(2, "0");
  return `${now.getFullYear()}-${mm}`;
}

interface Props {
  onSelect?: (group: IddirGroupData) => void;
  onPayPremium?: (groupId: string, premiumMinor: number, period: string, groupName: string) => void;
  onFileClaim?: (groupId: string, groupName: string) => void;
}

export function IddirGroupList({ onSelect, onPayPremium, onFileClaim }: Props = {}) {
  const { data, isLoading } = useQuery({
    queryKey: ["iddir-groups"],
    queryFn: fetchGroups,
  });

  const groups = data?.data ?? [];

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.primary} />
      </View>
    );
  }

  if (groups.length === 0) {
    return (
      <View style={styles.center}>
        <Text style={styles.empty}>No Iddir groups yet.</Text>
      </View>
    );
  }

  const period = currentPeriod();

  return (
    <View style={styles.list}>
      {groups.map((item) => {
        const isActive = item.coverage_status === "ACTIVE";
        return (
          <TouchableOpacity
            key={item.group_id}
            style={[styles.card, shadow.sm]}
            onPress={() => onSelect?.(item)}
            activeOpacity={onSelect ? 0.7 : 1}
            accessibilityRole={onSelect ? "button" : "none"}
          >
            <View style={styles.cardTop}>
              <View style={styles.cardInfo}>
                <Text style={styles.name}>{item.name}</Text>
                <Text style={styles.meta}>
                  Premium: {(item.premium_minor / 100).toFixed(0)} ETB · {item.frequency}
                </Text>
                <Text style={styles.benefit}>
                  Benefit: {(item.benefit_minor / 100).toFixed(0)} ETB
                </Text>
                <View style={[styles.badge, isActive ? styles.badgeActive : styles.badgeInactive]}>
                  <Text style={styles.badgeText}>{item.coverage_status}</Text>
                </View>
              </View>
            </View>
            {(onPayPremium || onFileClaim) && (
              <View style={styles.actions}>
                {onPayPremium && (
                  <TouchableOpacity
                    style={[styles.actionBtn, styles.premiumBtn]}
                    onPress={() => onPayPremium(item.group_id, item.premium_minor, period, item.name)}
                    accessibilityRole="button"
                    accessibilityLabel={`Pay premium for ${item.name}`}
                  >
                    <Text style={styles.actionBtnText}>Pay Premium</Text>
                  </TouchableOpacity>
                )}
                {onFileClaim && (
                  <TouchableOpacity
                    style={[styles.actionBtn, styles.claimBtn]}
                    onPress={() => onFileClaim(item.group_id, item.name)}
                    accessibilityRole="button"
                    accessibilityLabel={`File claim for ${item.name}`}
                  >
                    <Text style={styles.actionBtnText}>File Claim</Text>
                  </TouchableOpacity>
                )}
              </View>
            )}
          </TouchableOpacity>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  list: { padding: spacing.md, gap: spacing.sm },
  center: { paddingVertical: spacing.xl, alignItems: "center" },
  card: {
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    padding: spacing.md,
  },
  cardTop: { flexDirection: "row", alignItems: "flex-start" },
  cardInfo: { flex: 1 },
  name: { fontSize: 16, fontWeight: "600", color: colors.navy, marginBottom: 4 },
  meta: { fontSize: 13, color: colors.textSecondary, marginBottom: 4 },
  benefit: { fontSize: 13, color: colors.success, marginBottom: 8 },
  badge: { alignSelf: "flex-start", borderRadius: 6, paddingHorizontal: 8, paddingVertical: 2 },
  badgeActive: { backgroundColor: colors.successLight },
  badgeInactive: { backgroundColor: colors.errorLight },
  badgeText: { fontSize: 11, fontWeight: "600" },
  empty: { fontSize: 13, color: colors.textTertiary },
  actions: { flexDirection: "row", gap: spacing.sm, marginTop: spacing.sm },
  actionBtn: { borderRadius: radius.sm, paddingHorizontal: 14, paddingVertical: 7 },
  premiumBtn: { backgroundColor: colors.success },
  claimBtn: { backgroundColor: colors.error },
  actionBtnText: { fontSize: 12, fontWeight: "700", color: "#FFF" },
});
