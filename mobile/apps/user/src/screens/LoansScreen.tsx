import { useState } from "react";
import { View, Text, TouchableOpacity, ScrollView, StyleSheet } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import { LoanEligibilityCard } from "../../../../src/features/loans/LoanEligibilityCard";
import { LoanApplyForm } from "../../../../src/features/loans/LoanApplyForm";
import { ActiveLoanList } from "../../../../src/features/loans/ActiveLoanList";
import { RepayForm } from "../../../../src/features/loans/RepayForm";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AppNavigator";

type View_ = "list" | "apply" | "repay";

export function LoansScreen() {
  const [view, setView] = useState<View_>("list");
  const [repayTarget, setRepayTarget] = useState<{ loanId: string; outstandingMinor: number } | null>(null);

  const handleRepay = (loanId: string, outstandingMinor: number) => {
    setRepayTarget({ loanId, outstandingMinor });
    setView("repay");
  };

  if (view === "apply") {
    return (
      <SafeAreaView style={styles.safe}>
        <View style={styles.subHeader}>
          <TouchableOpacity style={styles.backBtn} onPress={() => setView("list")} accessibilityRole="button">
            <HugeiconsIcon icon={ArrowLeft01Icon} size={20} color={colors.text} strokeWidth={2} />
          </TouchableOpacity>
          <Text style={styles.subTitle}>Apply for Loan</Text>
          <View style={styles.backBtn} />
        </View>
        <ScrollView
          contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          <LoanApplyForm />
        </ScrollView>
      </SafeAreaView>
    );
  }

  if (view === "repay" && repayTarget) {
    return (
      <SafeAreaView style={styles.safe}>
        <View style={styles.subHeader}>
          <TouchableOpacity style={styles.backBtn} onPress={() => setView("list")} accessibilityRole="button">
            <HugeiconsIcon icon={ArrowLeft01Icon} size={20} color={colors.text} strokeWidth={2} />
          </TouchableOpacity>
          <Text style={styles.subTitle}>Repay Loan</Text>
          <View style={styles.backBtn} />
        </View>
        <ScrollView
          contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          <RepayForm
            loanId={repayTarget.loanId}
            outstandingMinor={repayTarget.outstandingMinor}
          />
        </ScrollView>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.pageHeader}>
        <Text style={styles.pageTitle}>Loans</Text>
        <Text style={styles.pageSubtitle}>Check eligibility and manage active loans</Text>
      </View>
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
        showsVerticalScrollIndicator={false}
      >
        <LoanEligibilityCard onApply={() => setView("apply")} />
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionLabel}>ACTIVE LOANS</Text>
        </View>
        <View style={[styles.card, shadow.sm]}>
          <ActiveLoanList onRepay={handleRepay} />
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
  pageHeader: { paddingHorizontal: spacing.md, paddingTop: spacing.md, marginBottom: spacing.md },
  pageTitle: { ...typography.title },
  pageSubtitle: { fontSize: 14, color: colors.textSecondary, marginTop: spacing.xs },
  content: { paddingHorizontal: spacing.md },
  sectionHeader: { marginTop: spacing.lg, marginBottom: spacing.sm },
  sectionLabel: { ...typography.label },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },
  subHeader: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
  },
  backBtn: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.surface, alignItems: "center", justifyContent: "center",
    shadowColor: "#000", shadowOpacity: 0.06, shadowRadius: 4, elevation: 2,
  },
  subTitle: { fontSize: 17, fontWeight: "700", color: colors.text },
});
