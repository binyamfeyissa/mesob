import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";

interface RegisterScreenProps {
  onSuccess: (registrationId: string) => void;
}

export function RegisterScreen({ onSuccess }: RegisterScreenProps) {
  const [msisdn, setMsisdn] = useState("");
  const [loading, setLoading] = useState(false);

  const handleRegister = async () => {
    if (!msisdn.match(/^\+2519\d{8}$/)) {
      Alert.alert("Invalid number", "Enter a valid Ethiopian mobile number (+2519XXXXXXXX)");
      return;
    }
    setLoading(true);
    try {
      const res = await apiFetch<{ data: { registration_id: string; otp_channel: string; expires_in: number } }>(
        "/identity/register",
        { method: "POST", body: JSON.stringify({ msisdn }) }
      );
      onSuccess(res.data.registration_id);
    } catch (e: any) {
      Alert.alert("Error", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>መሶብ ዋሌት / Mesob Wallet</Text>
      <Text style={styles.subtitle}>Register with your phone number</Text>
      <Text style={styles.label}>Phone Number (MSISDN)</Text>
      <TextInput
        style={styles.input}
        value={msisdn}
        onChangeText={setMsisdn}
        keyboardType="phone-pad"
        placeholder="+2519XXXXXXXX"
        accessibilityLabel="Ethiopian mobile phone number"
      />
      <TouchableOpacity
        style={[styles.btn, loading && styles.btnDisabled]}
        onPress={handleRegister}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Register"
      >
        <Text style={styles.btnText}>{loading ? "Sending OTP..." : "Register"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#0F1B3D", padding: 24, justifyContent: "center" },
  title: { color: "#E6A817", fontSize: 24, fontWeight: "700", textAlign: "center", marginBottom: 8 },
  subtitle: { color: "rgba(255,255,255,0.7)", fontSize: 14, textAlign: "center", marginBottom: 32 },
  label: { color: "rgba(255,255,255,0.8)", fontSize: 13, marginBottom: 6 },
  input: {
    backgroundColor: "rgba(255,255,255,0.1)", borderRadius: 10, paddingHorizontal: 14,
    paddingVertical: 14, fontSize: 16, color: "#FFF", marginBottom: 20, borderWidth: 1, borderColor: "rgba(255,255,255,0.2)"
  },
  btn: { backgroundColor: "#1B4FDE", borderRadius: 10, paddingVertical: 16, alignItems: "center" },
  btnDisabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
