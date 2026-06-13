import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ScrollView } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { verifyPin } from "../../utils/pin";

export function LoanApplyForm() {
  const [amountETB, setAmountETB] = useState("");
  const [termDays, setTermDays] = useState("30");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);

  const handleApply = async () => {
    const amount = parseFloat(amountETB);
    if (isNaN(amount) || amount <= 0 || !termDays || !pin) {
      Alert.alert("Fill in all fields including your PIN");
      return;
    }
    const pinOk = await verifyPin(pin);
    if (!pinOk) {
      Alert.alert("Incorrect PIN", "Please check your PIN and try again.");
      return;
    }
    setLoading(true);
    try {
      const res = await apiFetch<{
        data: {
          decision: string;
          loan_id?: string;
          principal_minor?: number;
          fee_minor?: number;
          due_date?: string;
          reasons?: string[];
        };
      }>(
        "/loans/apply",
        {
          method: "POST",
          headers: { "Idempotency-Key": generateCaptureKey() },
          body: JSON.stringify({
            amount_minor: Math.round(amount * 100),
            term_days: parseInt(termDays),
            idempotency_key: generateCaptureKey(),
          }),
        }
      );
      const d = res.data;
      if (d.decision === "APPROVED") {
        Alert.alert(
          "Loan Approved!",
          `Principal: ${((d.principal_minor ?? 0) / 100).toFixed(2)} ETB\nFee: ${((d.fee_minor ?? 0) / 100).toFixed(2)} ETB\nDue: ${d.due_date}`
        );
      } else {
        Alert.alert("Not approved", `Reasons: ${(d.reasons ?? []).join(", ")}`);
      }
      setAmountETB(""); setPin("");
    } catch (e: any) {
      Alert.alert("Application failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <ScrollView contentContainerStyle={styles.form}>
      <Text style={styles.title}>Apply for a Loan</Text>
      <Text style={styles.label}>Amount (ETB)</Text>
      <TextInput style={styles.input} value={amountETB} onChangeText={setAmountETB} keyboardType="decimal-pad" placeholder="0.00" accessibilityLabel="Loan amount" />
      <Text style={styles.label}>Term (days)</Text>
      <TextInput style={styles.input} value={termDays} onChangeText={setTermDays} keyboardType="number-pad" placeholder="30" accessibilityLabel="Loan term in days" />
      <Text style={styles.label}>Your PIN</Text>
      <TextInput style={styles.input} value={pin} onChangeText={setPin} keyboardType="number-pad" secureTextEntry maxLength={6} placeholder="••••" accessibilityLabel="PIN" />
      <TouchableOpacity style={[styles.btn, loading && styles.disabled]} onPress={handleApply} disabled={loading} accessibilityRole="button">
        <Text style={styles.btnText}>{loading ? "Applying..." : "Apply"}</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  form: { padding: 16 },
  title: { fontSize: 20, fontWeight: "700", color: "#0F1B3D", marginBottom: 20 },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 6 },
  input: { borderWidth: 1, borderColor: "#D1D5DB", borderRadius: 10, paddingHorizontal: 14, paddingVertical: 14, fontSize: 16, marginBottom: 14, backgroundColor: "#FFF" },
  btn: { backgroundColor: "#1B4FDE", borderRadius: 10, paddingVertical: 16, alignItems: "center", marginTop: 8 },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
