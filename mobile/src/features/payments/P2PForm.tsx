import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";
import { verifyPin } from "../../utils/pin";

export function P2PForm() {
  const [toMsisdn, setToMsisdn] = useState("");
  const [amountETB, setAmountETB] = useState("");
  const [note, setNote] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSend = async () => {
    const amount = parseFloat(amountETB);
    if (!toMsisdn || isNaN(amount) || amount <= 0 || !pin) {
      Alert.alert("Fill in all fields including your PIN");
      return;
    }
    const pinOk = await verifyPin(pin);
    if (!pinOk) {
      Alert.alert("Incorrect PIN", "Please check your PIN and try again.");
      return;
    }
    setLoading(true);
    const idempotencyKey = generateCaptureKey();
    try {
      const res = await apiFetch<{ data: { transaction_id: string; status: string } }>(
        "/payments/p2p",
        {
          method: "POST",
          body: JSON.stringify({
            to_msisdn: toMsisdn,
            amount_minor: Math.round(amount * 100),
            note,
            idempotency_key: idempotencyKey,
          }),
        }
      );
      Alert.alert("Sent", `${amountETB} ETB sent to ${toMsisdn}.\nTxn: ${res.data.transaction_id}`);
      setToMsisdn(""); setAmountETB(""); setNote(""); setPin("");
    } catch (e: any) {
      Alert.alert("Transfer failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.form}>
      <Text style={styles.title}>Send Money</Text>
      <Text style={styles.label}>To (Phone Number)</Text>
      <TextInput style={styles.input} value={toMsisdn} onChangeText={setToMsisdn} keyboardType="phone-pad" placeholder="+2519XXXXXXXX" accessibilityLabel="Recipient phone number" />
      <Text style={styles.label}>Amount (ETB)</Text>
      <TextInput style={styles.input} value={amountETB} onChangeText={setAmountETB} keyboardType="decimal-pad" placeholder="0.00" accessibilityLabel="Amount in ETB" />
      <Text style={styles.label}>Note (optional)</Text>
      <TextInput style={styles.input} value={note} onChangeText={setNote} placeholder="What's this for?" accessibilityLabel="Transfer note" />
      <Text style={styles.label}>Your PIN</Text>
      <TextInput style={styles.input} value={pin} onChangeText={setPin} keyboardType="number-pad" secureTextEntry maxLength={6} placeholder="••••" accessibilityLabel="Your PIN" />
      <TouchableOpacity style={[styles.btn, loading && styles.disabled]} onPress={handleSend} disabled={loading} accessibilityRole="button" accessibilityLabel="Send money">
        <Text style={styles.btnText}>{loading ? "Sending..." : "Send"}</Text>
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
