import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ActivityIndicator } from "react-native";
import { useState } from "react";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow, typography } from "../../theme/tokens";

const AGENT_PRIMARY = "#16A34A";

export function OnboardForm() {
  const [fan, setFan] = useState("");
  const [businessName, setBusinessName] = useState("");
  const [regionId, setRegionId] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!fan.trim() || !regionId.trim()) {
      Alert.alert("Required fields", "Fayda Account Number and Region ID are required.");
      return;
    }
    setSubmitting(true);
    try {
      await apiFetch("/agent/onboard", {
        method: "POST",
        body: JSON.stringify({
          fan: fan.trim(),
          region_id: regionId.trim(),
          business_name: businessName.trim() || undefined,
        }),
      });
      Alert.alert(
        "Application Submitted",
        "Your agent onboarding request has been submitted for branch review. You will be notified once approved."
      );
      setFan(""); setBusinessName(""); setRegionId("");
    } catch (e: unknown) {
      Alert.alert("Error", e instanceof Error ? e.message : "Try again");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <View style={styles.wrapper}>
      <View style={[styles.card, shadow.sm]}>
        <Text style={styles.title}>Agent Onboarding</Text>
        <Text style={styles.subtitle}>Submit your application to become a Mesob agent</Text>

        <Text style={styles.label}>Fayda Account Number (FAN)*</Text>
        <View style={styles.inputWrapper}>
          <TextInput
            style={styles.input}
            value={fan}
            onChangeText={setFan}
            placeholder="Your national ID number"
            placeholderTextColor={colors.textTertiary}
            autoCapitalize="none"
            accessibilityLabel="Fayda Account Number"
          />
        </View>

        <Text style={[styles.label, { marginTop: spacing.md }]}>Region ID*</Text>
        <View style={styles.inputWrapper}>
          <TextInput
            style={styles.input}
            value={regionId}
            onChangeText={setRegionId}
            placeholder="UUID of your region"
            placeholderTextColor={colors.textTertiary}
            autoCapitalize="none"
            accessibilityLabel="Region ID"
          />
        </View>

        <Text style={[styles.label, { marginTop: spacing.md }]}>Business Name (optional)</Text>
        <View style={styles.inputWrapper}>
          <TextInput
            style={styles.input}
            value={businessName}
            onChangeText={setBusinessName}
            placeholder="e.g. Haile General Store"
            placeholderTextColor={colors.textTertiary}
            accessibilityLabel="Business name"
          />
        </View>

        <TouchableOpacity
          style={[styles.btn, submitting && styles.btnDisabled]}
          onPress={handleSubmit}
          disabled={submitting}
          accessibilityRole="button"
        >
          {submitting ? <ActivityIndicator color="#FFF" /> : <Text style={styles.btnText}>Submit Application</Text>}
        </TouchableOpacity>

        <Text style={styles.hint}>
          A branch officer will review your application and approve your float account.
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { paddingHorizontal: spacing.md },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, padding: spacing.lg },
  title: { ...typography.heading, marginBottom: 4 },
  subtitle: { fontSize: 13, color: colors.textSecondary, marginBottom: spacing.xl },
  label: { ...typography.captionBold, color: colors.textSecondary, marginBottom: spacing.sm },
  inputWrapper: {
    borderWidth: 1.5, borderColor: colors.border, borderRadius: radius.md,
    backgroundColor: colors.surface, paddingHorizontal: spacing.md, height: 54, justifyContent: "center",
  },
  input: { fontSize: 16, color: colors.text },
  btn: {
    backgroundColor: AGENT_PRIMARY, borderRadius: radius.md,
    height: 54, alignItems: "center", justifyContent: "center", marginTop: spacing.lg, ...shadow.md,
  },
  btnDisabled: { opacity: 0.55 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
  hint: { fontSize: 12, color: colors.textTertiary, textAlign: "center", marginTop: spacing.md, lineHeight: 18 },
});
