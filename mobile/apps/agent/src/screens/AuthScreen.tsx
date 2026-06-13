import { useState } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  ActivityIndicator,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { CallIcon, LockIcon, EyeIcon, EyeOffIcon } from "@hugeicons/core-free-icons";
import { apiFetch } from "../../../../src/api/client";
import { useAuthStore } from "../../../../src/lib/authStore";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";

const AGENT_PRIMARY = "#16A34A";
const AGENT_PRIMARY_LIGHT = "rgba(22,163,74,0.12)";
const DEMO_AGENT = { msisdn: "+251911000004", pin: "111111" };

export function AuthScreen({ onAuthenticated }: { onAuthenticated: () => void }) {
  const { setSession } = useAuthStore();
  const [msisdn, setMsisdn] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);
  const [pinVisible, setPinVisible] = useState(false);

  const doLogin = async (m: string, p: string) => {
    setLoading(true);
    try {
      const res = await apiFetch<{
        data: { access_token: string; refresh_token: string; role: string; kyc_tier: number; user_id: string };
      }>("/identity/login", { method: "POST", body: JSON.stringify({ msisdn: m, pin: p }) });

      if (res.data.role !== "AGENT") {
        Alert.alert("Access Denied", "This app is for agents only.");
        return;
      }

      setSession({
        accessToken: res.data.access_token,
        refreshToken: res.data.refresh_token,
        role: res.data.role,
        userId: res.data.user_id,
        kycTier: res.data.kyc_tier,
        msisdn: m,
      });

      // Fetch agent profile to cache region_id for OnboardForm
      try {
        const profile = await apiFetch<{ data: { region_id?: string } }>("/agent/float");
        if (profile.data.region_id) {
          await AsyncStorage.setItem("agent_region_id", profile.data.region_id);
        }
      } catch {
        // Non-fatal — OnboardForm falls back to default region
      }

      onAuthenticated();
    } catch (e: unknown) {
      Alert.alert("Login failed", e instanceof Error ? e.message : "Try again");
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.safe}>
      <KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : "height"} style={styles.flex}>
        <ScrollView
          contentContainerStyle={styles.container}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          {/* Brand mark */}
          <View style={styles.brandRow}>
            <View style={styles.brandIcon}>
              <Text style={styles.brandLetter}>M</Text>
            </View>
          </View>

          <Text style={styles.heading}>Mesob Agent</Text>
          <Text style={styles.subheading}>For registered agents only</Text>

          <View style={styles.fieldGroup}>
            <Text style={styles.fieldLabel}>Phone Number</Text>
            <View style={styles.inputWrapper}>
              <HugeiconsIcon icon={CallIcon} size={18} color={colors.textTertiary} strokeWidth={1.8} style={styles.inputIcon} />
              <TextInput
                style={styles.input}
                value={msisdn}
                onChangeText={setMsisdn}
                keyboardType="phone-pad"
                placeholder="+2519XXXXXXXX"
                placeholderTextColor={colors.textTertiary}
                autoComplete="tel"
                accessibilityLabel="Phone number"
              />
            </View>

            <Text style={[styles.fieldLabel, { marginTop: spacing.md }]}>PIN</Text>
            <View style={styles.inputWrapper}>
              <HugeiconsIcon icon={LockIcon} size={18} color={colors.textTertiary} strokeWidth={1.8} style={styles.inputIcon} />
              <TextInput
                style={[styles.input, styles.inputFlex]}
                value={pin}
                onChangeText={setPin}
                keyboardType="number-pad"
                secureTextEntry={!pinVisible}
                maxLength={6}
                placeholder="••••••"
                placeholderTextColor={colors.textTertiary}
                accessibilityLabel="PIN"
              />
              <TouchableOpacity onPress={() => setPinVisible((v) => !v)} style={styles.eyeBtn}>
                <HugeiconsIcon
                  icon={pinVisible ? EyeIcon : EyeOffIcon}
                  size={18}
                  color={colors.textTertiary}
                  strokeWidth={1.8}
                />
              </TouchableOpacity>
            </View>
          </View>

          <TouchableOpacity
            style={[styles.btn, loading && styles.btnDisabled]}
            onPress={() => {
              if (!msisdn || !pin) { Alert.alert("Enter phone and PIN"); return; }
              doLogin(msisdn, pin);
            }}
            disabled={loading}
            accessibilityRole="button"
          >
            {loading ? <ActivityIndicator color="#FFF" /> : <Text style={styles.btnText}>Log In as Agent</Text>}
          </TouchableOpacity>

          <View style={styles.demoSection}>
            <Text style={styles.demoLabel}>Demo Agent Account</Text>
            <TouchableOpacity
              style={styles.demoBtn}
              onPress={() => doLogin(DEMO_AGENT.msisdn, DEMO_AGENT.pin)}
              disabled={loading}
              accessibilityRole="button"
            >
              <Text style={styles.demoBtnText}>{DEMO_AGENT.msisdn}</Text>
              <Text style={styles.demoBtnSub}>PIN: {DEMO_AGENT.pin}</Text>
            </TouchableOpacity>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.surface },
  flex: { flex: 1 },
  container: { flexGrow: 1, paddingHorizontal: spacing.lg, paddingTop: spacing.xl, paddingBottom: spacing.xxl },
  brandRow: { marginBottom: spacing.lg },
  brandIcon: {
    width: 52, height: 52, borderRadius: 14,
    backgroundColor: AGENT_PRIMARY, alignItems: "center", justifyContent: "center",
  },
  brandLetter: { fontSize: 22, fontWeight: "800", color: "#FFF" },
  heading: { ...typography.title, marginBottom: spacing.xs },
  subheading: { ...typography.body, color: colors.textSecondary, marginBottom: spacing.xl },
  fieldGroup: { marginBottom: spacing.lg },
  fieldLabel: { ...typography.captionBold, color: colors.textSecondary, marginBottom: spacing.sm },
  inputWrapper: {
    flexDirection: "row", alignItems: "center",
    backgroundColor: colors.surface, borderRadius: radius.md,
    borderWidth: 1.5, borderColor: colors.border,
    paddingHorizontal: spacing.md, height: 54,
  },
  inputIcon: { marginRight: spacing.sm },
  input: { flex: 1, fontSize: 16, color: colors.text, height: "100%" },
  inputFlex: { flex: 1 },
  eyeBtn: { padding: spacing.xs },
  btn: {
    backgroundColor: AGENT_PRIMARY, borderRadius: radius.md,
    height: 54, alignItems: "center", justifyContent: "center", ...shadow.md,
  },
  btnDisabled: { opacity: 0.55 },
  btnText: { color: "#FFF", fontSize: 16, fontWeight: "700" },
  demoSection: {
    marginTop: spacing.xl, paddingTop: spacing.lg,
    borderTopWidth: 1, borderTopColor: colors.divider, alignItems: "center",
  },
  demoLabel: { ...typography.label, marginBottom: spacing.md },
  demoBtn: {
    backgroundColor: AGENT_PRIMARY_LIGHT, borderRadius: radius.md,
    paddingHorizontal: spacing.xl, paddingVertical: spacing.md, alignItems: "center",
  },
  demoBtnText: { fontSize: 13, fontWeight: "600", color: AGENT_PRIMARY },
  demoBtnSub: { fontSize: 11, color: colors.textSecondary, marginTop: 2 },
});
