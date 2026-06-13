import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";

interface OTPScreenProps {
  registrationId: string;
  onSuccess: (challengeToken: string) => void;
}

export function OTPScreen({ registrationId, onSuccess }: OTPScreenProps) {
  const [otp, setOtp] = useState("");
  const [loading, setLoading] = useState(false);

  const handleVerify = async () => {
    if (otp.length < 4) { Alert.alert("Enter the OTP you received"); return; }
    setLoading(true);
    try {
      const res = await apiFetch<{ data: { verified: boolean; challenge_token: string } }>(
        "/identity/verify-otp",
        { method: "POST", body: JSON.stringify({ registration_id: registrationId, otp }) }
      );
      if (!res.data.verified) { Alert.alert("Wrong OTP", "Try again"); return; }
      onSuccess(res.data.challenge_token);
    } catch (e: any) {
      Alert.alert("Error", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Verify OTP</Text>
      <Text style={styles.subtitle}>Enter the code sent to your phone</Text>
      <TextInput
        style={styles.input}
        value={otp}
        onChangeText={setOtp}
        keyboardType="number-pad"
        maxLength={6}
        placeholder="------"
        placeholderTextColor="rgba(255,255,255,0.3)"
        accessibilityLabel="One-time password"
        textAlign="center"
      />
      <TouchableOpacity
        style={[styles.btn, loading && styles.btnDisabled]}
        onPress={handleVerify}
        disabled={loading}
        accessibilityRole="button"
      >
        <Text style={styles.btnText}>{loading ? "Verifying..." : "Verify"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#0F1B3D", padding: 24, justifyContent: "center" },
  title: { color: "#FFF", fontSize: 22, fontWeight: "700", textAlign: "center", marginBottom: 8 },
  subtitle: { color: "rgba(255,255,255,0.6)", fontSize: 14, textAlign: "center", marginBottom: 32 },
  input: {
    backgroundColor: "rgba(255,255,255,0.1)", borderRadius: 10, paddingHorizontal: 14,
    paddingVertical: 18, fontSize: 28, color: "#FFF", marginBottom: 24, letterSpacing: 12,
    borderWidth: 1, borderColor: "rgba(255,255,255,0.2)"
  },
  btn: { backgroundColor: "#1B4FDE", borderRadius: 10, paddingVertical: 16, alignItems: "center" },
  btnDisabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
