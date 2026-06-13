import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ActivityIndicator } from "react-native";
import { useState } from "react";
import { generateCaptureKey } from "../../sync/idempotency";
import { operationQueue } from "../../sync/queue";
import { apiFetch, getAccessToken } from "../../api/client";
import { colors, spacing, radius, shadow, typography } from "../../theme/tokens";

const AGENT_PRIMARY = "#16A34A";

export function CashInForm() {
  const [msisdn, setMsisdn] = useState("");
  const [amountETB, setAmountETB] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    const amount = parseFloat(amountETB);
    if (!msisdn || isNaN(amount) || amount <= 0) {
      Alert.alert("Invalid input", "Enter a valid phone number and amount.");
      return;
    }
    setLoading(true);
    const captureKey = generateCaptureKey();
    const amountMinor = Math.round(amount * 100);

    try {
      if (getAccessToken()) {
        await apiFetch("/agent/cash-in", {
          method: "POST",
          body: JSON.stringify({
            customer_msisdn: msisdn,
            amount_minor: amountMinor,
            idempotency_key: captureKey,
          }),
        });
        Alert.alert("Success", `Cash-in of ${amountETB} ETB for ${msisdn} completed.`);
      } else {
        operationQueue.enqueue({
          id: captureKey,
          type: "CASH_IN",
          payload: { customer_msisdn: msisdn, amount_minor: amountMinor },
          capturedAt: new Date().toISOString(),
        });
        Alert.alert("Saved Offline", `Cash-in for ${msisdn} saved. Sync when back online.`);
      }
      setMsisdn("");
      setAmountETB("");
    } catch (e: unknown) {
      operationQueue.enqueue({
        id: captureKey,
        type: "CASH_IN",
        payload: { customer_msisdn: msisdn, amount_minor: amountMinor },
        capturedAt: new Date().toISOString(),
      });
      Alert.alert("Saved Offline", `Couldn't reach server. Cash-in saved for sync.`);
      setMsisdn("");
      setAmountETB("");
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.wrapper}>
      <View style={[styles.card, shadow.sm]}>
        <Text style={styles.title}>Cash In</Text>
        <Text style={styles.subtitle}>Deposit cash into a customer wallet</Text>

        <Text style={styles.label}>Customer Phone</Text>
        <View style={styles.inputWrapper}>
          <TextInput
            style={styles.input}
            value={msisdn}
            onChangeText={setMsisdn}
            keyboardType="phone-pad"
            placeholder="+2519XXXXXXXX"
            placeholderTextColor={colors.textTertiary}
            accessibilityLabel="Customer phone number"
          />
        </View>

        <Text style={[styles.label, { marginTop: spacing.md }]}>Amount (ETB)</Text>
        <View style={styles.inputWrapper}>
          <TextInput
            style={styles.input}
            value={amountETB}
            onChangeText={setAmountETB}
            keyboardType="decimal-pad"
            placeholder="0.00"
            placeholderTextColor={colors.textTertiary}
            accessibilityLabel="Amount in Ethiopian Birr"
          />
        </View>

        <TouchableOpacity
          style={[styles.btn, loading && styles.btnDisabled]}
          onPress={handleSubmit}
          disabled={loading}
          accessibilityRole="button"
        >
          {loading ? <ActivityIndicator color="#FFF" /> : <Text style={styles.btnText}>Cash In</Text>}
        </TouchableOpacity>
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
});
