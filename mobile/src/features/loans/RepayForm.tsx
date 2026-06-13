import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { verifyPin } from "../../utils/pin";

interface RepayFormProps {
  loanId: string;
  outstandingMinor: number;
}

export function RepayForm({ loanId, outstandingMinor }: RepayFormProps) {
  const [amountETB, setAmountETB] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);

  const handleRepay = async () => {
    const amount = parseFloat(amountETB);
    if (isNaN(amount) || amount <= 0 || !pin) {
      Alert.alert("Enter amount and PIN");
      return;
    }
    const pinOk = await verifyPin(pin);
    if (!pinOk) {
      Alert.alert("Incorrect PIN", "Please check your PIN and try again.");
      return;
    }
    setLoading(true);
    try {
      const res = await apiFetch<{ data: { loan_id: string; outstanding_minor: number; status: string } }>(
        `/loans/${loanId}/repay`,
        {
          method: "POST",
          body: JSON.stringify({
            amount_minor: Math.round(amount * 100),
            idempotency_key: generateCaptureKey(),
          }),
        }
      );
      const d = res.data;
      Alert.alert(
        d.status === "REPAID" ? "Loan fully repaid!" : "Payment recorded",
        `Outstanding: ${(d.outstanding_minor / 100).toFixed(2)} ETB`
      );
      setAmountETB(""); setPin("");
    } catch (e: any) {
      Alert.alert("Repayment failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.card}>
      <Text style={styles.title}>Repay Loan</Text>
      <Text style={styles.outstanding}>Outstanding: {(outstandingMinor / 100).toFixed(2)} ETB</Text>
      <Text style={styles.label}>Amount to Repay (ETB)</Text>
      <TextInput style={styles.input} value={amountETB} onChangeText={setAmountETB} keyboardType="decimal-pad" placeholder="0.00" accessibilityLabel="Repayment amount" />
      <Text style={styles.label}>Your PIN</Text>
      <TextInput style={styles.input} value={pin} onChangeText={setPin} keyboardType="number-pad" secureTextEntry maxLength={6} placeholder="••••" accessibilityLabel="PIN" />
      <TouchableOpacity style={[styles.btn, loading && styles.disabled]} onPress={handleRepay} disabled={loading} accessibilityRole="button">
        <Text style={styles.btnText}>{loading ? "Processing..." : "Repay"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  card: { margin: 16, backgroundColor: "#FFF", borderRadius: 16, padding: 20 },
  title: { fontSize: 18, fontWeight: "700", color: "#0F1B3D", marginBottom: 4 },
  outstanding: { fontSize: 13, color: "#DC2626", marginBottom: 20 },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 6 },
  input: { borderWidth: 1, borderColor: "#D1D5DB", borderRadius: 10, paddingHorizontal: 14, paddingVertical: 14, fontSize: 16, marginBottom: 14, backgroundColor: "#F9FAFB" },
  btn: { backgroundColor: "#059669", borderRadius: 10, paddingVertical: 16, alignItems: "center" },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
