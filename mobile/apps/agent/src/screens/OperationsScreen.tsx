import { useState } from "react";
import { View, Text, TouchableOpacity, ScrollView, StyleSheet } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { CashInForm } from "../../../../src/features/cashops/CashInForm";
import { CashOutForm } from "../../../../src/features/cashops/CashOutForm";
import { colors, spacing, radius, shadow } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AgentNavigator";

type OpTab = "cashin" | "cashout";

const AGENT_PRIMARY = "#16A34A";

export function OperationsScreen() {
  const [activeTab, setActiveTab] = useState<OpTab>("cashin");

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.tabRow}>
        {(["cashin", "cashout"] as OpTab[]).map((key) => {
          const isActive = activeTab === key;
          const label = key === "cashin" ? "Cash In" : "Cash Out";
          return (
            <TouchableOpacity
              key={key}
              style={[styles.tabPill, isActive && styles.tabPillActive]}
              onPress={() => setActiveTab(key)}
              accessibilityRole="tab"
              accessibilityState={{ selected: isActive }}
            >
              <Text style={[styles.tabLabel, isActive && styles.tabLabelActive]}>{label}</Text>
            </TouchableOpacity>
          );
        })}
      </View>
      <ScrollView
        contentContainerStyle={{ paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16, paddingTop: spacing.md }}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        {activeTab === "cashin" ? <CashInForm /> : <CashOutForm />}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
  tabRow: {
    flexDirection: "row",
    marginHorizontal: spacing.md,
    marginTop: spacing.md,
    marginBottom: spacing.xs,
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    padding: 4,
    ...shadow.sm,
  },
  tabPill: { flex: 1, paddingVertical: 10, borderRadius: radius.sm, alignItems: "center" },
  tabPillActive: { backgroundColor: AGENT_PRIMARY },
  tabLabel: { fontSize: 14, fontWeight: "600", color: colors.textSecondary },
  tabLabelActive: { color: "#FFF" },
});
