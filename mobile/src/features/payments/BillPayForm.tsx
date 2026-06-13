import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { verifyPin } from "../../utils/pin";

export function BillPayForm() {
  const [billerId, setBillerId] = useState("");
  const [accountRef, setAccountRef] = useState("");
  const [amountETB, setAmountETB] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);

  const handlePay = async () => {
    const amount = parseFloat(amountETB);
    if (!billerId || !accountRef || isNaN(amount) || amount <= 0 || !pin) {
      Alert.alert("Fill in all fields");
      return;
    }
    const pinOk = await verifyPin(pin);
    if (!pinOk) {
      Alert.alert("Incorrect PIN", "Please check your PIN and try again.");
      return;
    }
    setLoading(true);
    try {
      // Bill payment holds PENDING until biller confirms — never optimistically completed
      const res = await apiFetch<{ data: { transaction_id: string; status: string; biller_ref: string } }>(
        "/payments/bill",
        {
          method: "POST",
          body: JSON.stringify({
            biller_id: billerId,
            reference: accountRef,
            amount_minor: Math.round(amount * 100),
            idempotency_key: generateCaptureKey(),
          }),
        }
      );
      const msg = res.data.status === "COMPLETED"
        ? `Bill paid. Ref: ${res.data.biller_ref}`
        : `Payment submitted. Status: PENDING — awaiting biller confirmation.`;
      Alert.alert("Done", msg);
      setBillerId(""); setAccountRef(""); setAmountETB(""); setPin("");
    } catch (e: any) {
      Alert.alert("Payment failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.form}>
      <Text style={styles.title}>Pay Bill</Text>
      <Text style={styles.note}>Status stays PENDING until the biller confirms receipt.</Text>
      <Text style={styles.label}>Biller ID</Text>
      <TextInput style={styles.input} value={billerId} onChangeText={setBillerId} placeholder="Biller ID" accessibilityLabel="Biller ID" />
      <Text style={styles.label}>Account / Bill Reference</Text>
      <TextInput style={styles.input} value={accountRef} onChangeText={setAccountRef} placeholder="Account number" accessibilityLabel="Bill account reference" />
      <Text style={styles.label}>Amount (ETB)</Text>
      <TextInput style={styles.input} value={amountETB} onChangeText={setAmountETB} keyboardType="decimal-pad" placeholder="0.00" accessibilityLabel="Amount" />
      <Text style={styles.label}>Your PIN</Text>
      <TextInput style={styles.input} value={pin} onChangeText={setPin} keyboardType="number-pad" secureTextEntry maxLength={6} placeholder="••••" accessibilityLabel="PIN" />
      <TouchableOpacity style={[styles.btn, loading && styles.disabled]} onPress={handlePay} disabled={loading} accessibilityRole="button">
        <Text style={styles.btnText}>{loading ? "Submitting..." : "Pay Bill"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  form: { padding: 16 },
  title: { fontSize: 20, fontWeight: "700", color: "#0F1B3D", marginBottom: 8 },
  note: { fontSize: 12, color: "#6B7280", marginBottom: 20, fontStyle: "italic" },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 6 },
  input: { borderWidth: 1, borderColor: "#D1D5DB", borderRadius: 10, paddingHorizontal: 14, paddingVertical: 14, fontSize: 16, marginBottom: 14, backgroundColor: "#FFF" },
  btn: { backgroundColor: "#1B4FDE", borderRadius: 10, paddingVertical: 16, alignItems: "center", marginTop: 8 },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
