import { TouchableOpacity, Text, StyleSheet, ActivityIndicator } from "react-native";
import { colors, radius, shadow } from "../theme/tokens";

interface PrimaryButtonProps {
  label: string;
  onPress: () => void;
  loading?: boolean;
  disabled?: boolean;
  variant?: "primary" | "dark" | "success" | "danger";
}

export function PrimaryButton({
  label,
  onPress,
  loading = false,
  disabled = false,
  variant = "primary",
}: PrimaryButtonProps) {
  const bgMap = {
    primary: colors.primary,
    dark: colors.navy,
    success: colors.success,
    danger: colors.error,
  };

  return (
    <TouchableOpacity
      style={[
        styles.btn,
        { backgroundColor: bgMap[variant] },
        (disabled || loading) && styles.disabled,
        shadow.md,
      ]}
      onPress={onPress}
      disabled={disabled || loading}
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityState={{ disabled: disabled || loading }}
    >
      {loading ? (
        <ActivityIndicator color="#FFF" />
      ) : (
        <Text style={styles.label}>{label}</Text>
      )}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  btn: {
    borderRadius: radius.md,
    paddingVertical: 16,
    minHeight: 54,
    alignItems: "center",
    justifyContent: "center",
  },
  disabled: { opacity: 0.45 },
  label: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
