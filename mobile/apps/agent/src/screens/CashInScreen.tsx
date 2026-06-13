import { ScrollView, StyleSheet } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { CashInForm } from "../../../../src/features/cashops/CashInForm";
import { colors, spacing } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AgentNavigator";

export function CashInScreen() {
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        contentContainerStyle={{ paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16, paddingTop: spacing.md }}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        <CashInForm />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
});
