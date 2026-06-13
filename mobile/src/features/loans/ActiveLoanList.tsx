import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius } from "../../theme/tokens";

interface Loan {
  loan_id: string;
  status: string;
  outstanding_minor: number;
  due_date: string;
}

async function fetchLoans(): Promise<{ data: Loan[] }> {
  return apiFetch("/loans");
}

interface Props {
  onRepay?: (loanId: string, outstandingMinor: number) => void;
}

export function ActiveLoanList({ onRepay }: Props = {}) {
  const queryClient = useQueryClient();
  const { data } = useQuery({ queryKey: ["loans"], queryFn: fetchLoans });
  const loans = data?.data ?? [];

  if (loans.length === 0) {
    return (
      <View style={styles.empty}>
        <Text style={styles.emptyText}>No active loans.</Text>
      </View>
    );
  }

  return (
    <View style={styles.list}>
      {loans.map((item) => {
        const isActive = item.status === "ACTIVE";
        return (
          <View key={item.loan_id} style={styles.row}>
            <View style={styles.rowInfo}>
              <Text style={styles.amount}>{(item.outstanding_minor / 100).toFixed(2)} ETB outstanding</Text>
              <Text style={styles.due}>Due: {item.due_date}</Text>
            </View>
            <View style={styles.rowRight}>
              <View style={[styles.badge, isActive ? styles.active : styles.overdue]}>
                <Text style={styles.badgeText}>{item.status}</Text>
              </View>
              {isActive && onRepay && (
                <TouchableOpacity
                  style={styles.repayBtn}
                  onPress={() => onRepay(item.loan_id, item.outstanding_minor)}
                  accessibilityRole="button"
                >
                  <Text style={styles.repayBtnText}>Repay</Text>
                </TouchableOpacity>
              )}
            </View>
          </View>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  list: { padding: spacing.md, gap: spacing.sm },
  row: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    backgroundColor: colors.surface,
    borderRadius: radius.sm,
    padding: 14,
  },
  rowInfo: { flex: 1 },
  amount: { fontSize: 15, fontWeight: "600", color: colors.navy },
  due: { fontSize: 12, color: colors.textTertiary, marginTop: 2 },
  rowRight: { alignItems: "flex-end", gap: 6 },
  badge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 3 },
  active: { backgroundColor: colors.successLight },
  overdue: { backgroundColor: colors.errorLight },
  badgeText: { fontSize: 11, fontWeight: "600" },
  repayBtn: {
    backgroundColor: colors.success,
    borderRadius: 6,
    paddingHorizontal: 12,
    paddingVertical: 5,
  },
  repayBtnText: { fontSize: 12, fontWeight: "700", color: "#FFF" },
  empty: { paddingVertical: spacing.lg, paddingHorizontal: spacing.md, alignItems: "center" },
  emptyText: { fontSize: 13, color: colors.textTertiary },
});
