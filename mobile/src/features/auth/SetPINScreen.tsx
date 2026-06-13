import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { apiFetch } from "../../api/client";
import { useAuthStore } from "../../lib/authStore";
import { storePin } from "../../utils/pin";

interface SetPINScreenProps {
  challengeToken: string;
  onSuccess: () => void;
}

export function SetPINScreen({ challengeToken, onSuccess }: SetPINScreenProps) {
  const { setSession } = useAuthStore();
  const [pin, setPin] = useState("");
  const [confirm, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSetPIN = async () => {
    if (pin.length < 4) { Alert.alert("PIN must be at least 4 digits"); return; }
    if (pin !== confirm) { Alert.alert("PINs do not match"); return; }
    setLoading(true);
    try {
      const res = await apiFetch<{
        data: { access_token: string; refresh_token: string; user_id: string; role: string; kyc_tier: number; wallet_account_id?: string };
      }>("/identity/set-pin", {
        method: "POST",
        body: JSON.stringify({ challenge_token: challengeToken, pin }),
      });
      await storePin(pin);
      setSession({
        accessToken: res.data.access_token,
        refreshToken: res.data.refresh_token,
        role: res.data.role,
        userId: res.data.user_id,
        kycTier: res.data.kyc_tier,
        msisdn: "",
        walletAccountId: res.data.wallet_account_id ?? "",
      });
      onSuccess();
    } catch (e: unknown) {
      Alert.alert("Error", e instanceof Error ? e.message : "Failed to set PIN");
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Set Your PIN</Text>
      <Text style={styles.subtitle}>Choose a 4-6 digit PIN. Do not share it with anyone.</Text>
      <Text style={styles.label}>PIN</Text>
      <TextInput
        style={styles.input}
        value={pin}
        onChangeText={setPin}
        keyboardType="number-pad"
        secureTextEntry
        maxLength={6}
        placeholder="••••"
        placeholderTextColor="rgba(255,255,255,0.3)"
        accessibilityLabel="PIN"
      />
      <Text style={styles.label}>Confirm PIN</Text>
      <TextInput
        style={styles.input}
        value={confirm}
        onChangeText={setConfirm}
        keyboardType="number-pad"
        secureTextEntry
        maxLength={6}
        placeholder="••••"
        placeholderTextColor="rgba(255,255,255,0.3)"
        accessibilityLabel="Confirm PIN"
      />
      <TouchableOpacity
        style={[styles.btn, loading && styles.btnDisabled]}
        onPress={handleSetPIN}
        disabled={loading}
        accessibilityRole="button"
      >
        <Text style={styles.btnText}>{loading ? "Setting PIN..." : "Set PIN & Continue"}</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#0F1B3D", padding: 24, justifyContent: "center" },
  title: { color: "#FFF", fontSize: 22, fontWeight: "700", textAlign: "center", marginBottom: 8 },
  subtitle: { color: "rgba(255,255,255,0.6)", fontSize: 13, textAlign: "center", marginBottom: 32 },
  label: { color: "rgba(255,255,255,0.8)", fontSize: 13, marginBottom: 6 },
  input: {
    backgroundColor: "rgba(255,255,255,0.1)", borderRadius: 10, paddingHorizontal: 14,
    paddingVertical: 14, fontSize: 24, color: "#FFF", marginBottom: 16,
    borderWidth: 1, borderColor: "rgba(255,255,255,0.2)", letterSpacing: 8
  },
  btn: { backgroundColor: "#1B4FDE", borderRadius: 10, paddingVertical: 16, alignItems: "center", marginTop: 8 },
  btnDisabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
