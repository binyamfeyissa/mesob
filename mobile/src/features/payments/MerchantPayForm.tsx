import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { verifyPin } from "../../utils/pin";

export function MerchantPayForm() {
  const [merchantId, setMerchantId] = useState("");
  const [amountETB, setAmountETB] = useState("");
  const [ref, setRef] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);

  const handlePay = async () => {
    const amount = parseFloat(amountETB);
    if (!merchantId || isNaN(amount) || amount <= 0 || !pin) {
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
      const res = await apiFetch<{ data: { transaction_id: string; status: string; receipt_ref: string } }>(
        "/payments/merchant",
        {
          method: "POST",
          body: JSON.stringify({
            merchant_id: merchantId,
            amount_minor: Math.round(amount * 100),
            idempotency_key: generateCaptureKey(),
          }),
        }
      );
      Alert.alert("Paid", `Receipt: ${res.data.receipt_ref}`);
      setMerchantId(""); setAmountETB(""); setRef(""); setPin("");
    } catch (e: any) {
      Alert.alert("Payment failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.form}>
      <Text style={styles.title}>Pay Merchant</Text>
      <Text style={styles.label}>Merchant ID</Text>
      <TextInput style={styles.input} value={merchantId} onChangeText={setMerchantId} placeholder="Merchant ID" accessibilityLabel="Merchant ID" />
      <Text style={styles.label}>Amount (ETB)</Text>
      <TextInput style={styles.input} value={amountETB} onChangeText={setAmountETB} keyboardType="decimal-pad" placeholder="0.00" accessibilityLabel="Amount" />
      <Text style={styles.label}>Reference / Order #</Text>
      <TextInput style={styles.input} value={ref} onChangeText={setRef} placeholder="Order reference" accessibilityLabel="Order reference" />
      <Text style={styles.label}>Your PIN</Text>
      <TextInput style={styles.input} value={pin} onChangeText={setPin} keyboardType="number-pad" secureTextEntry maxLength={6} placeholder="••••" accessibilityLabel="PIN" />
      <TouchableOpacity style={[styles.btn, loading && styles.disabled]} onPress={handlePay} disabled={loading} accessibilityRole="button">
        <Text style={styles.btnText}>{loading ? "Processing..." : "Pay"}</Text>
      </TouchableOpacity>
    </View>
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
