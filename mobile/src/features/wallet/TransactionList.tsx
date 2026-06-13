import { View, Text, StyleSheet } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { useAuthStore } from "../../lib/authStore";
import { colors, spacing, radius } from "../../theme/tokens";

interface Entry {
  id: string;
  direction: "D" | "C";
  amount_minor: number;
  currency?: string;
  counterparty?: string;
  created_at: string;
}

interface EntriesResponse {
  data: Entry[];
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-ET", { day: "numeric", month: "short", year: "numeric" });
}

function EntryIcon({ label, isCredit }: { label: string; isCredit: boolean }) {
  const initials = label
    .split(" ")
    .map((w) => w[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);

  return (
    <View style={[styles.iconCircle, isCredit && styles.iconCircleCredit]}>
      <Text style={[styles.iconInitials, isCredit && { color: colors.success }]}>
        {initials || (isCredit ? "+" : "−")}
      </Text>
    </View>
  );
}

export function TransactionList() {
  const user = useAuthStore((s) => s.user);
  const accountId = user?.walletAccountId ?? "";

  const { data } = useQuery<EntriesResponse>({
    queryKey: ["entries", accountId],
    queryFn: () => apiFetch(`/ledger/accounts/${accountId}/entries?limit=20`),
    enabled: !!accountId,
  });

  const entries = data?.data ?? [];

  if (entries.length === 0) {
    return (
      <View style={styles.empty}>
        <Text style={styles.emptyText}>No transactions yet</Text>
      </View>
    );
  }

  return (
    <View>
      {entries.map((item, index) => {
        const isCredit = item.direction === "C";
        const amount = (item.amount_minor / 100).toFixed(2);
        const label = item.counterparty ?? (isCredit ? "Received" : "Sent");
        const isLast = index === entries.length - 1;

        return (
          <View key={item.id}>
            <View style={styles.row}>
              <EntryIcon label={label} isCredit={isCredit} />
              <View style={styles.rowMeta}>
                <Text style={styles.rowTitle}>{label}</Text>
                <Text style={styles.rowDate}>{formatDate(item.created_at)}</Text>
              </View>
              <Text style={[styles.rowAmount, isCredit ? styles.amountCredit : styles.amountDebit]}>
                {isCredit ? "+" : "−"}{amount} ETB
              </Text>
            </View>
            {!isLast && <View style={styles.separator} />}
          </View>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 14,
    paddingHorizontal: spacing.md,
  },

  iconCircle: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.bg,
    alignItems: "center",
    justifyContent: "center",
    marginRight: spacing.md,
  },
  iconCircleCredit: {
    backgroundColor: colors.successLight,
  },
  iconInitials: {
    fontSize: 14,
    fontWeight: "700",
    color: colors.textSecondary,
  },

  rowMeta: { flex: 1 },
  rowTitle: { fontSize: 15, fontWeight: "600", color: colors.text },
  rowDate: { fontSize: 12, color: colors.textTertiary, marginTop: 2 },

  rowAmount: { fontSize: 15, fontWeight: "700" },
  amountCredit: { color: colors.success },
  amountDebit: { color: colors.text },

  separator: {
    height: 1,
    backgroundColor: colors.divider,
    marginLeft: 44 + spacing.md + spacing.md,
  },

  empty: { paddingVertical: spacing.xl, alignItems: "center" },
  emptyText: { fontSize: 13, color: colors.textTertiary },
});
