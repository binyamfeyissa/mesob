import { useState } from "react";
import { View, Text, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";

interface PayPremiumFormProps {
  groupId: string;
  premiumMinor: number;
  period: string; // e.g. "2026-05"
}

export function PayPremiumForm({ groupId, premiumMinor, period }: PayPremiumFormProps) {
  const [loading, setLoading] = useState(false);

  const handlePay = async () => {
    setLoading(true);
    try {
      const res = await apiFetch<{ data: { premium_id: string; transaction_id: string; coverage: string } }>(
        `/iddir/groups/${groupId}/premium`,
        {
          method: "POST",
          body: JSON.stringify({ period, idempotency_key: generateCaptureKey() }),
        }
      );
      Alert.alert("Premium paid", `Coverage: ${res.data.coverage}`);
    } catch (e: any) {
      Alert.alert("Payment failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.card}>
      <Text style={styles.label}>Monthly Premium — {period}</Text>
      <Text style={styles.amount}>{(premiumMinor / 100).toFixed(2)} ETB</Text>
      <TouchableOpacity
        style={[styles.btn, loading && styles.disabled]}
        onPress={handlePay}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Pay Iddir premium"
      >
        <Text style={styles.btnText}>{loading ? "Processing..." : "Pay Premium"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  card: { margin: 16, backgroundColor: "#FFF", borderRadius: 16, padding: 20 },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 4 },
  amount: { fontSize: 28, fontWeight: "700", color: "#0F1B3D", marginBottom: 20 },
  btn: { backgroundColor: "#059669", borderRadius: 10, paddingVertical: 16, alignItems: "center" },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
