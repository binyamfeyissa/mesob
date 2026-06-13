import { useState } from "react";
import { ScrollView, View, Text, TouchableOpacity, StyleSheet, SafeAreaView } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { P2PForm } from "../../../../src/features/payments/P2PForm";
import { BillPayForm } from "../../../../src/features/payments/BillPayForm";
import { MerchantPayForm } from "../../../../src/features/payments/MerchantPayForm";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AppNavigator";

type PayTab = "send" | "bills" | "merchant";

const PAY_TABS: { key: PayTab; icon: keyof typeof Ionicons.glyphMap; label: string }[] = [
  { key: "send", icon: "paper-plane-outline", label: "Send Money" },
  { key: "bills", icon: "receipt-outline", label: "Pay Bills" },
  { key: "merchant", icon: "storefront-outline", label: "Merchant" },
];

export function PaymentsScreen() {
  const [activeTab, setActiveTab] = useState<PayTab>("send");

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
        showsVerticalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
      >
        {/* Header */}
        <View style={styles.pageHeader}>
          <Text style={styles.pageTitle}>Payments</Text>
        </View>

        {/* Tab switcher */}
        <View style={[styles.tabRow, shadow.sm]}>
          {PAY_TABS.map((t) => {
            const active = activeTab === t.key;
            return (
              <TouchableOpacity
                key={t.key}
                style={[styles.tabBtn, active && styles.tabBtnActive]}
                onPress={() => setActiveTab(t.key)}
                accessibilityRole="tab"
                accessibilityState={{ selected: active }}
              >
                <Ionicons
                  name={t.icon}
                  size={18}
                  color={active ? colors.primary : colors.textTertiary}
                />
                <Text style={[styles.tabBtnText, active && styles.tabBtnTextActive]}>
                  {t.label}
                </Text>
              </TouchableOpacity>
            );
          })}
        </View>

        {/* Form content */}
        <View style={[styles.formCard, shadow.sm]}>
          {activeTab === "send" && <P2PForm />}
          {activeTab === "bills" && <BillPayForm />}
          {activeTab === "merchant" && <MerchantPayForm />}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
  content: { paddingHorizontal: spacing.md, paddingTop: spacing.md },

  pageHeader: { marginBottom: spacing.lg },
  pageTitle: { ...typography.title },

  tabRow: {
    flexDirection: "row",
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    padding: 6,
    marginBottom: spacing.md,
    gap: 4,
  },
  tabBtn: {
    flex: 1,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: spacing.xs,
    paddingVertical: 10,
    borderRadius: radius.md,
  },
  tabBtnActive: {
    backgroundColor: colors.primaryLight,
  },
  tabBtnText: {
    fontSize: 12,
    fontWeight: "500",
    color: colors.textTertiary,
  },
  tabBtnTextActive: {
    color: colors.primary,
    fontWeight: "600",
  },

  formCard: {
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    overflow: "hidden",
  },
});
