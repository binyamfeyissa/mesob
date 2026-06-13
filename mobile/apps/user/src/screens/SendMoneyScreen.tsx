import { View, Text, TouchableOpacity, ScrollView, StyleSheet } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import { P2PForm } from "../../../../src/features/payments/P2PForm";
import { useNav } from "../navigation/NavContext";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";

export function SendMoneyScreen() {
  const { goBack } = useNav();
  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.header}>
        <TouchableOpacity style={styles.backBtn} onPress={goBack} accessibilityRole="button" accessibilityLabel="Back">
          <HugeiconsIcon icon={ArrowLeft01Icon} size={20} color={colors.text} strokeWidth={2} />
        </TouchableOpacity>
        <Text style={styles.title}>Send Money</Text>
        <View style={styles.backBtn} />
      </View>
      <ScrollView
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        <View style={[styles.card, shadow.sm]}>
          <P2PForm />
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
  },
  backBtn: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.surface, alignItems: "center", justifyContent: "center", ...shadow.sm,
  },
  title: { ...typography.subheading },
  content: { padding: spacing.md },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },
});
