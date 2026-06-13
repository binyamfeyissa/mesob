import { View, Text, TouchableOpacity, StyleSheet, Alert, ActivityIndicator } from "react-native";
import { useState } from "react";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import { generateCaptureKey } from "../../sync/idempotency";
import { operationQueue } from "../../sync/queue";
import { apiFetch, getAccessToken } from "../../api/client";
import { colors, spacing, radius, shadow, typography } from "../../theme/tokens";

const AGENT_PRIMARY = "#16A34A";

export function CashOutForm() {
  const [step, setStep] = useState<1 | 2>(1);
  const [msisdn, setMsisdn] = useState("");
  const [amountETB, setAmountETB] = useState("");
  const [loading, setLoading] = useState(false);
  const [captureKey] = useState(() => generateCaptureKey());

  const handleNext = () => {
    const amount = parseFloat(amountETB);
    if (!msisdn) { Alert.alert("Enter customer phone number"); return; }
    if (isNaN(amount) || amount <= 0) { Alert.alert("Enter a valid amount"); return; }
    setStep(2);
  };

  const handleConfirm = async () => {
    setLoading(true);
    const amountMinor = Math.round(parseFloat(amountETB) * 100);

    try {
      if (getAccessToken()) {
        await apiFetch("/agent/cash-out", {
          method: "POST",
          body: JSON.stringify({
            customer_msisdn: msisdn,
            amount_minor: amountMinor,
            idempotency_key: captureKey,
          }),
        });
        Alert.alert("Success", `Cash-out of ${amountETB} ETB for ${msisdn} completed.`);
      } else {
        operationQueue.enqueue({
          id: captureKey,
          type: "CASH_OUT",
          payload: { customer_msisdn: msisdn, amount_minor: amountMinor },
          capturedAt: new Date().toISOString(),
        });
        Alert.alert("Saved Offline", `Cash-out saved. Sync when back online.`);
      }
      setMsisdn(""); setAmountETB(""); setStep(1);
    } catch {
      operationQueue.enqueue({
        id: captureKey,
        type: "CASH_OUT",
        payload: { customer_msisdn: msisdn, amount_minor: amountMinor },
        capturedAt: new Date().toISOString(),
      });
      Alert.alert("Saved Offline", `Couldn't reach server. Cash-out saved for sync.`);
      setMsisdn(""); setAmountETB(""); setStep(1);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.wrapper}>
      {step === 1 ? (
        <View style={[styles.card, shadow.sm]}>
          <Text style={styles.title}>Cash Out</Text>
          <Text style={styles.subtitle}>Withdraw cash from a customer wallet</Text>

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

          <TouchableOpacity style={styles.btn} onPress={handleNext} accessibilityRole="button">
            <Text style={styles.btnText}>Next →</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <View style={[styles.card, shadow.sm]}>
          <TouchableOpacity style={styles.backRow} onPress={() => setStep(1)} accessibilityRole="button">
            <HugeiconsIcon icon={ArrowLeft01Icon} size={18} color={AGENT_PRIMARY} strokeWidth={2} />
            <Text style={styles.backText}>Back</Text>
          </TouchableOpacity>

          <Text style={styles.title}>Confirm Cash Out</Text>
          <Text style={styles.subtitle}>
            Verify details with the customer before confirming.
          </Text>

          <View style={styles.summaryRow}>
            <Text style={styles.summaryLabel}>Phone</Text>
            <Text style={styles.summaryValue}>{msisdn}</Text>
          </View>
          <View style={[styles.summaryRow, styles.summaryBorder]}>
            <Text style={styles.summaryLabel}>Amount</Text>
            <Text style={styles.summaryValue}>{amountETB} ETB</Text>
          </View>

          <TouchableOpacity
            style={[styles.btn, { marginTop: spacing.lg }, loading && styles.btnDisabled]}
            onPress={handleConfirm}
            disabled={loading}
            accessibilityRole="button"
          >
            {loading ? <ActivityIndicator color="#FFF" /> : <Text style={styles.btnText}>Confirm Cash Out</Text>}
          </TouchableOpacity>
        </View>
      )}
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
  backRow: { flexDirection: "row", alignItems: "center", gap: spacing.xs, marginBottom: spacing.md },
  backText: { color: AGENT_PRIMARY, fontWeight: "600", fontSize: 14 },
  summaryRow: { flexDirection: "row", justifyContent: "space-between", paddingVertical: spacing.sm },
  summaryBorder: { borderTopWidth: 1, borderTopColor: colors.divider },
  summaryLabel: { fontSize: 13, color: colors.textSecondary },
  summaryValue: { fontSize: 13, fontWeight: "600", color: colors.text },
});
