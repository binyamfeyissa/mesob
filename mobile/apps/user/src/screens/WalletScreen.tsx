import { ScrollView, View, Text, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import type { IconSvgElement } from "@hugeicons/react-native";
import {
  MoneySend01Icon,
  ReceiptDollarIcon,
  Store01Icon,
  Notification01Icon,
} from "@hugeicons/core-free-icons";
import { BalanceCard } from "../../../../src/features/wallet/BalanceCard";
import { TransactionList } from "../../../../src/features/wallet/TransactionList";
import { useAuthStore } from "../../../../src/lib/authStore";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AppNavigator";
import { useNav, SubScreen } from "../navigation/NavContext";

function initials(msisdn: string) {
  return msisdn ? msisdn.slice(-2) : "ME";
}

function ServiceButton({
  icon,
  label,
  color,
  onPress,
}: {
  icon: IconSvgElement;
  label: string;
  color: string;
  onPress: () => void;
}) {
  return (
    <TouchableOpacity style={styles.serviceBtn} onPress={onPress} accessibilityRole="button">
      <View style={[styles.serviceIcon, { backgroundColor: color + "18" }]}>
        <HugeiconsIcon icon={icon} size={22} color={color} strokeWidth={1.8} />
      </View>
      <Text style={styles.serviceBtnLabel}>{label}</Text>
    </TouchableOpacity>
  );
}

export function WalletScreen() {
  const { user } = useAuthStore();
  const { push } = useNav();

  const services: { icon: IconSvgElement; label: string; color: string; screen: SubScreen }[] = [
    { icon: MoneySend01Icon, label: "Send Money", color: colors.primary, screen: "send-money" },
    { icon: ReceiptDollarIcon, label: "Pay Bills", color: colors.success, screen: "pay-bills" },
    { icon: Store01Icon, label: "Merchant", color: colors.warning, screen: "merchant" },
  ];

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.avatarCircle}>
            <Text style={styles.avatarText}>{initials(user?.msisdn ?? "")}</Text>
          </View>
          <TouchableOpacity style={styles.iconBtn} accessibilityRole="button" accessibilityLabel="Notifications">
            <HugeiconsIcon icon={Notification01Icon} size={20} color={colors.text} strokeWidth={1.8} />
          </TouchableOpacity>
        </View>

        {/* Balance */}
        <View style={[styles.card, shadow.sm, { marginBottom: spacing.lg }]}>
          <BalanceCard
            onSend={() => push("send-money")}
            onPayBills={() => push("pay-bills")}
            onAddMoney={() =>
              Alert.alert(
                "Add Money",
                "To add money to your wallet, visit any Mesob Agent near you. They will perform a Cash In on your behalf.",
                [{ text: "OK" }]
              )
            }
          />
        </View>

        {/* Services */}
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionLabel}>SERVICES</Text>
        </View>
        <View style={[styles.card, shadow.sm, styles.servicesCard]}>
          {services.map((s) => (
            <ServiceButton
              key={s.screen}
              icon={s.icon}
              label={s.label}
              color={s.color}
              onPress={() => push(s.screen)}
            />
          ))}
        </View>

        {/* Recent Transactions */}
        <View style={[styles.sectionHeader, { marginTop: spacing.lg }]}>
          <Text style={styles.sectionLabel}>RECENT TRANSACTIONS</Text>
          <TouchableOpacity accessibilityRole="button">
            <Text style={styles.seeAll}>See all</Text>
          </TouchableOpacity>
        </View>
        <View style={[styles.card, shadow.sm]}>
          <TransactionList />
        </View>
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
    justifyContent: "space-between",
    marginBottom: spacing.lg,
  },
  avatarCircle: {
    width: 42, height: 42, borderRadius: 21,
    backgroundColor: colors.bg, borderWidth: 1, borderColor: colors.border,
    alignItems: "center", justifyContent: "center",
  },
  avatarText: { fontSize: 14, fontWeight: "700", color: colors.textSecondary },
  iconBtn: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.surface, alignItems: "center", justifyContent: "center", ...shadow.sm,
  },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },
  servicesCard: {
    flexDirection: "row", paddingVertical: spacing.lg,
    paddingHorizontal: spacing.md, justifyContent: "space-around",
  },
  serviceBtn: { alignItems: "center", gap: spacing.sm },
  serviceIcon: { width: 56, height: 56, borderRadius: 28, alignItems: "center", justifyContent: "center" },
  serviceBtnLabel: { fontSize: 12, fontWeight: "500", color: colors.textSecondary },
  sectionHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: spacing.sm },
  sectionLabel: { ...typography.label },
  seeAll: { fontSize: 13, fontWeight: "600", color: colors.primary },
});
