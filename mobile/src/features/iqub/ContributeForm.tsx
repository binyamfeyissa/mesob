import { useState } from "react";
import { View, Text, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { generateCaptureKey } from "../../sync/idempotency";

interface ContributeFormProps {
  groupId: string;
  cycleId: string;
  cycleMinor: number;
}

export function ContributeForm({ groupId, cycleId, cycleMinor }: ContributeFormProps) {
  const [loading, setLoading] = useState(false);

  const handleContribute = async () => {
    setLoading(true);
    try {
      const res = await apiFetch<{ data: { contribution_id: string; transaction_id: string; cycle_status: string } }>(
        `/iqub/groups/${groupId}/contribute`,
        {
          method: "POST",
          body: JSON.stringify({ cycle_id: cycleId, idempotency_key: generateCaptureKey() }),
        }
      );
      Alert.alert("Contributed!", `Cycle status: ${res.data.cycle_status}`);
    } catch (e: any) {
      Alert.alert("Contribution failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.card}>
      <Text style={styles.label}>Your contribution this cycle</Text>
      <Text style={styles.amount}>{(cycleMinor / 100).toFixed(2)} ETB</Text>
      <TouchableOpacity
        style={[styles.btn, loading && styles.disabled]}
        onPress={handleContribute}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Contribute to Iqub cycle"
      >
        <Text style={styles.btnText}>{loading ? "Processing..." : "Contribute Now"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  card: { margin: 16, backgroundColor: "#FFF", borderRadius: 16, padding: 20 },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 4 },
  amount: { fontSize: 28, fontWeight: "700", color: "#0F1B3D", marginBottom: 20 },
  btn: { backgroundColor: "#1B4FDE", borderRadius: 10, paddingVertical: 16, alignItems: "center" },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
