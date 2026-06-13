import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";

interface FileClaimFormProps {
  groupId: string;
  onFiled: (claimId: string) => void;
}

export function FileClaimForm({ groupId, onFiled }: FileClaimFormProps) {
  const [type, setType] = useState("");
  const [description, setDescription] = useState("");
  const [evidenceRef, setEvidenceRef] = useState("");
  const [loading, setLoading] = useState(false);

  const handleFile = async () => {
    if (!type || !description) { Alert.alert("Fill in claim type and description"); return; }
    setLoading(true);
    try {
      const res = await apiFetch<{ data: { claim_id: string; status: string } }>(
        `/iddir/groups/${groupId}/claims`,
        { method: "POST", body: JSON.stringify({ type, description, evidence_ref: evidenceRef }) }
      );
      Alert.alert("Claim filed", `Status: ${res.data.status}\nID: ${res.data.claim_id}`);
      onFiled(res.data.claim_id);
    } catch (e: any) {
      Alert.alert("Failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.form}>
      <Text style={styles.title}>File a Claim</Text>
      <Text style={styles.label}>Claim Type (e.g. DEATH, ILLNESS)</Text>
      <TextInput style={styles.input} value={type} onChangeText={setType} autoCapitalize="characters" placeholder="DEATH" accessibilityLabel="Claim type" />
      <Text style={styles.label}>Description</Text>
      <TextInput style={[styles.input, styles.multiline]} value={description} onChangeText={setDescription} multiline numberOfLines={4} placeholder="Describe the event..." accessibilityLabel="Claim description" />
      <Text style={styles.label}>Evidence Reference (optional)</Text>
      <TextInput style={styles.input} value={evidenceRef} onChangeText={setEvidenceRef} placeholder="Document ID / photo ref" accessibilityLabel="Evidence reference" />
      <TouchableOpacity style={[styles.btn, loading && styles.disabled]} onPress={handleFile} disabled={loading} accessibilityRole="button">
        <Text style={styles.btnText}>{loading ? "Filing..." : "File Claim"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  form: { padding: 16 },
  title: { fontSize: 20, fontWeight: "700", color: "#0F1B3D", marginBottom: 20 },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 6 },
  input: { borderWidth: 1, borderColor: "#D1D5DB", borderRadius: 10, paddingHorizontal: 14, paddingVertical: 14, fontSize: 16, marginBottom: 14, backgroundColor: "#FFF" },
  multiline: { height: 100, textAlignVertical: "top" },
  btn: { backgroundColor: "#DC2626", borderRadius: 10, paddingVertical: 16, alignItems: "center", marginTop: 8 },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
