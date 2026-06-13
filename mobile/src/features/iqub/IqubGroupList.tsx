import { View, Text, TouchableOpacity, StyleSheet, ActivityIndicator } from "react-native";
import type { IqubGroupData } from "./IqubGroupDetail";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

interface IqubGroup {
  group_id: string;
  name: string;
  cycle_minor: number;
  frequency: string;
  cycle: { id: string; number: number; paid: number; total: number; next_payout_member: string; due_date: string };
}

async function fetchGroups(): Promise<{ data: IqubGroup[] }> {
  return apiFetch("/iqub/groups");
}

interface Props {
  onSelect?: (group: IqubGroupData) => void;
  onContribute?: (groupId: string, cycleId: string, cycleMinor: number, groupName: string) => void;
}

export function IqubGroupList({ onSelect, onContribute }: Props = {}) {
  const { data, isLoading } = useQuery({
    queryKey: ["iqub-groups"],
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
        <Text style={styles.empty}>No Iqub groups yet.</Text>
      </View>
    );
  }

  return (
    <View style={styles.list}>
      {groups.map((item) => (
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
                {(item.cycle_minor / 100).toFixed(0)} ETB · {item.frequency}
              </Text>
              <Text style={styles.cycle}>
                Cycle {item.cycle?.number}: {item.cycle?.paid}/{item.cycle?.total} paid
              </Text>
              <Text style={styles.due}>Due: {item.cycle?.due_date ?? "—"}</Text>
            </View>
            {onContribute && item.cycle?.id && (
              <TouchableOpacity
                style={styles.actionBtn}
                onPress={() => onContribute(item.group_id, item.cycle.id, item.cycle_minor, item.name)}
                accessibilityRole="button"
                accessibilityLabel={`Contribute to ${item.name}`}
              >
                <Text style={styles.actionBtnText}>Contribute</Text>
              </TouchableOpacity>
            )}
          </View>
        </TouchableOpacity>
      ))}
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
  cardTop: { flexDirection: "row", alignItems: "flex-start", justifyContent: "space-between" },
  cardInfo: { flex: 1, marginRight: spacing.sm },
  name: { fontSize: 16, fontWeight: "600", color: colors.navy, marginBottom: 4 },
  meta: { fontSize: 13, color: colors.textSecondary, marginBottom: 4 },
  cycle: { fontSize: 13, color: colors.primary },
  due: { fontSize: 12, color: colors.textTertiary, marginTop: 2 },
  empty: { fontSize: 13, color: colors.textTertiary },
  actionBtn: {
    backgroundColor: colors.primary,
    borderRadius: radius.sm,
    paddingHorizontal: 12,
    paddingVertical: 7,
    alignSelf: "flex-start",
    marginTop: 2,
  },
  actionBtnText: { fontSize: 12, fontWeight: "700", color: "#FFF" },
});
