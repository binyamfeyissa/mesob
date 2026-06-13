import { useState } from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { HugeiconsIcon } from "@hugeicons/react-native";
import type { IconSvgElement } from "@hugeicons/react-native";
import {
  Add01Icon,
  ArrowUp01Icon,
  ReceiptDollarIcon,
  ArrowDown01Icon,
  EyeIcon,
  EyeOffIcon,
} from "@hugeicons/core-free-icons";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { useAuthStore } from "../../lib/authStore";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

interface BalanceResponse {
  data: { account_id: string; balance_minor: number; currency: string; as_of: string };
}

function ActionButton({ icon, label, onPress }: { icon: IconSvgElement; label: string; onPress: () => void }) {
  return (
    <TouchableOpacity style={styles.actionBtn} onPress={onPress} accessibilityRole="button">
      <View style={styles.actionCircle}>
        <HugeiconsIcon icon={icon} size={20} color={colors.text} strokeWidth={1.8} />
      </View>
      <Text style={styles.actionLabel}>{label}</Text>
    </TouchableOpacity>
  );
}

interface BalanceCardProps {
  onSend?: () => void;
  onPayBills?: () => void;
  onAddMoney?: () => void;
}

export function BalanceCard({ onSend, onPayBills, onAddMoney }: BalanceCardProps = {}) {
  const { user } = useAuthStore();
  const [balanceVisible, setBalanceVisible] = useState(true);

  const { data } = useQuery<BalanceResponse>({
    queryKey: ["balance", user?.walletAccountId],
    queryFn: () => apiFetch(`/ledger/accounts/${user?.walletAccountId}/balance`),
    enabled: !!user?.walletAccountId,
    refetchInterval: 60_000,
  });

  const balance = data?.data?.balance_minor ?? 0;
  const formatted = (balance / 100).toLocaleString("en-ET", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  const [whole, decimal] = formatted.split(".");

  return (
    <View style={styles.container}>
      <View style={styles.balancePill}>
        <Text style={styles.balancePillText}>TOTAL BALANCE</Text>
        <HugeiconsIcon icon={ArrowDown01Icon} size={12} color={colors.textSecondary} strokeWidth={2} />
      </View>

      <View style={styles.amountRow}>
        <Text style={styles.currency}>ETB </Text>
        <Text style={styles.whole}>{balanceVisible ? whole : "••••"}</Text>
        {balanceVisible && <Text style={styles.decimal}>.{decimal}</Text>}
        <TouchableOpacity
          onPress={() => setBalanceVisible((v) => !v)}
          style={styles.eyeBtn}
          accessibilityRole="button"
          accessibilityLabel={balanceVisible ? "Hide balance" : "Show balance"}
        >
          <View style={styles.eyeCircle}>
            <HugeiconsIcon
              icon={balanceVisible ? EyeIcon : EyeOffIcon}
              size={16}
              color={colors.textSecondary}
              strokeWidth={1.8}
            />
          </View>
        </TouchableOpacity>
      </View>

      <View style={styles.actionsRow}>
        <ActionButton icon={Add01Icon} label="Add Money" onPress={onAddMoney ?? (() => {})} />
        <ActionButton icon={ArrowUp01Icon} label="Send" onPress={onSend ?? (() => {})} />
        <ActionButton icon={ReceiptDollarIcon} label="Pay Bills" onPress={onPayBills ?? (() => {})} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    paddingTop: spacing.lg,
    paddingBottom: spacing.lg,
    backgroundColor: colors.surface,
  },
  balancePill: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
    backgroundColor: colors.bg,
    borderRadius: radius.full,
    paddingHorizontal: 14,
    paddingVertical: 7,
    marginBottom: spacing.md,
  },
  balancePillText: { fontSize: 12, fontWeight: "600", color: colors.textSecondary, letterSpacing: 0.5 },
  amountRow: { flexDirection: "row", alignItems: "flex-end", marginBottom: spacing.lg },
  currency: { fontSize: 22, fontWeight: "700", color: colors.textSecondary, paddingBottom: 6 },
  whole: { fontSize: 48, fontWeight: "800", color: colors.text, lineHeight: 54 },
  decimal: { fontSize: 24, fontWeight: "700", color: colors.text, paddingBottom: 6 },
  eyeBtn: { marginLeft: spacing.sm, paddingBottom: 4 },
  eyeCircle: {
    width: 32, height: 32, borderRadius: 16,
    backgroundColor: colors.bg, alignItems: "center", justifyContent: "center",
  },
  actionsRow: { flexDirection: "row", justifyContent: "center", gap: 32 },
  actionBtn: { alignItems: "center", gap: 8 },
  actionCircle: {
    width: 56, height: 56, borderRadius: 28,
    backgroundColor: colors.bg, alignItems: "center", justifyContent: "center",
    ...shadow.sm,
  },
  actionLabel: { fontSize: 12, fontWeight: "500", color: colors.textSecondary },
});
