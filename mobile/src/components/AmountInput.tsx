import { View, Text, TextInput, StyleSheet } from "react-native";
import { colors, spacing, radius, typography } from "../theme/tokens";

interface AmountInputProps {
  value: string;
  onChangeText: (text: string) => void;
  label?: string;
  currency?: string;
  accessibilityLabel?: string;
}

export function AmountInput({
  value,
  onChangeText,
  label = "Amount",
  currency = "ETB",
  accessibilityLabel,
}: AmountInputProps) {
  return (
    <View style={styles.wrapper}>
      {label && <Text style={styles.label}>{label}</Text>}
      <View style={styles.row}>
        <Text style={styles.currency}>{currency}</Text>
        <TextInput
          style={styles.input}
          value={value}
          onChangeText={onChangeText}
          keyboardType="decimal-pad"
          placeholder="0.00"
          placeholderTextColor={colors.textTertiary}
          accessibilityLabel={accessibilityLabel ?? label}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { marginBottom: spacing.md },
  label: { ...typography.captionBold, marginBottom: spacing.sm },
  row: {
    flexDirection: "row",
    alignItems: "center",
    borderWidth: 1.5,
    borderColor: colors.border,
    borderRadius: radius.md,
    backgroundColor: colors.surface,
    paddingHorizontal: spacing.md,
    height: 54,
  },
  currency: {
    fontSize: 15,
    color: colors.textSecondary,
    marginRight: spacing.sm,
    fontWeight: "600",
  },
  input: {
    flex: 1,
    fontSize: 20,
    color: colors.text,
    fontWeight: "600",
  },
});
