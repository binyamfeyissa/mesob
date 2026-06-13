import { useState } from "react";
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
import { apiFetch } from "../../api/client";
import { useAuthStore } from "../../lib/authStore";
import { colors, spacing, radius, typography, shadow } from "../../theme/tokens";

interface LoginScreenProps {
  onSuccess: () => void;
  onRegister: () => void;
}

const DEMO_ACCOUNTS = [
  { label: "User (T2)", msisdn: "+251911000005", pin: "111111" },
  { label: "Agent", msisdn: "+251911000004", pin: "111111" },
];

export function LoginScreen({ onSuccess, onRegister }: LoginScreenProps) {
  const { setSession } = useAuthStore();
  const [msisdn, setMsisdn] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);
  const [pinVisible, setPinVisible] = useState(false);

  const doLogin = async (m: string, p: string) => {
    setLoading(true);
    try {
      const res = await apiFetch<{
        data: { access_token: string; refresh_token: string; role: string; kyc_tier: number; user_id: string; wallet_account_id?: string };
      }>("/identity/login", { method: "POST", body: JSON.stringify({ msisdn: m, pin: p }) });
      setSession({
        accessToken: res.data.access_token,
        refreshToken: res.data.refresh_token,
        role: res.data.role,
        userId: res.data.user_id,
        kycTier: res.data.kyc_tier,
        msisdn: m,
        walletAccountId: res.data.wallet_account_id ?? "",
      });
      onSuccess();
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
          <View style={styles.avatarRow}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>M</Text>
            </View>
          </View>

          <Text style={styles.heading}>Welcome to{"\n"}Mesob Wallet</Text>
          <Text style={styles.subheading}>Sign in to your account</Text>

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
            {loading ? <ActivityIndicator color="#FFF" /> : <Text style={styles.btnText}>Log In</Text>}
          </TouchableOpacity>

          <TouchableOpacity onPress={onRegister} style={styles.registerRow} accessibilityRole="button">
            <Text style={styles.registerText}>
              New user? <Text style={styles.registerLink}>Register here</Text>
            </Text>
          </TouchableOpacity>

          <View style={styles.demoSection}>
            <Text style={styles.demoLabel}>Demo Accounts</Text>
            <View style={styles.demoRow}>
              {DEMO_ACCOUNTS.map((a) => (
                <TouchableOpacity
                  key={a.label}
                  style={styles.demoBtn}
                  onPress={() => doLogin(a.msisdn, a.pin)}
                  disabled={loading}
                  accessibilityRole="button"
                >
                  <Text style={styles.demoBtnText}>{a.label}</Text>
                  <Text style={styles.demoBtnSub}>PIN: {a.pin}</Text>
                </TouchableOpacity>
              ))}
            </View>
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
  avatarRow: { marginBottom: spacing.lg },
  avatar: {
    width: 52, height: 52, borderRadius: 26,
    backgroundColor: colors.bg, alignItems: "center", justifyContent: "center",
  },
  avatarText: { fontSize: 20, fontWeight: "700", color: colors.textSecondary },
  heading: { ...typography.title, marginBottom: spacing.xs, lineHeight: 36 },
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
    backgroundColor: colors.primary, borderRadius: radius.md,
    height: 54, alignItems: "center", justifyContent: "center", ...shadow.md,
  },
  btnDisabled: { opacity: 0.55 },
  btnText: { color: "#FFF", fontSize: 16, fontWeight: "700" },
  registerRow: { alignItems: "center", marginTop: spacing.lg },
  registerText: { fontSize: 14, color: colors.textSecondary },
  registerLink: { color: colors.primary, fontWeight: "600" },
  demoSection: {
    marginTop: spacing.xl, paddingTop: spacing.lg,
    borderTopWidth: 1, borderTopColor: colors.divider,
  },
  demoLabel: { ...typography.label, textAlign: "center", marginBottom: spacing.md },
  demoRow: { flexDirection: "row", justifyContent: "center", gap: 12 },
  demoBtn: {
    backgroundColor: colors.primaryLight, borderRadius: radius.md,
    paddingHorizontal: spacing.lg, paddingVertical: spacing.md, alignItems: "center",
  },
  demoBtnText: { fontSize: 13, fontWeight: "600", color: colors.primary },
  demoBtnSub: { fontSize: 11, color: colors.textSecondary, marginTop: 2 },
});
