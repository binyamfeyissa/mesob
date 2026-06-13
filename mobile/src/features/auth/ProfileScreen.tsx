import { useState } from "react";
import { ScrollView, View, Text, TouchableOpacity, StyleSheet, Alert, TextInput, ActivityIndicator } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import type { IconSvgElement } from "@hugeicons/react-native";
import {
  ArrowRight01Icon,
  ArrowLeft01Icon,
  UserIcon,
  ShieldUserIcon,
  Notification01Icon,
  ChartBarIncreasingIcon,
  LockIcon,
  GlobeIcon,
  HelpCircleIcon,
} from "@hugeicons/core-free-icons";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch, clearAccessToken } from "../../api/client";
import { useAuthStore } from "../../lib/authStore";
import { colors, spacing, radius, shadow, typography } from "../../theme/tokens";

const TAB_BAR_TOTAL_HEIGHT = 0;

interface ProfileResponse {
  data: {
    user_id: string;
    msisdn: string;
    kyc_tier: number;
    preferred_lang: string;
    fan_masked: string;
    status: string;
  };
}

const KYC_TIER_LABEL = ["Basic", "Standard", "Full KYC"];
const KYC_TIER_COLOR = [colors.warning, colors.primary, colors.success];
const LANG_OPTIONS: { key: string; label: string }[] = [
  { key: "am", label: "Amharic" },
  { key: "om", label: "Afaan Oromoo" },
  { key: "ti", label: "Tigrinya" },
  { key: "en", label: "English" },
];
const LANG_LABEL: Record<string, string> = { am: "Amharic", om: "Afaan Oromoo", ti: "Tigrinya", en: "English" };

type SubView = "main" | "pin" | "language" | "kyc";

function MenuItem({
  icon, label, value, last, danger, onPress,
}: {
  icon: IconSvgElement; label: string; value?: string; last?: boolean; danger?: boolean; onPress?: () => void;
}) {
  return (
    <TouchableOpacity
      style={[styles.menuItem, !last && styles.menuItemBorder]}
      onPress={onPress}
      activeOpacity={0.7}
      accessibilityRole="button"
    >
      <View style={styles.menuItemLeft}>
        <View style={styles.menuIconWrap}>
          <HugeiconsIcon icon={icon} size={18} color={danger ? colors.error : colors.primary} strokeWidth={1.8} />
        </View>
        <Text style={[styles.menuItemLabel, danger && styles.menuItemDanger]}>{label}</Text>
      </View>
      <View style={styles.menuItemRight}>
        {value ? (
          <Text style={styles.menuItemValue}>{value}</Text>
        ) : (
          <HugeiconsIcon icon={ArrowRight01Icon} size={16} color={colors.textTertiary} strokeWidth={1.8} />
        )}
      </View>
    </TouchableOpacity>
  );
}

function SubHeader({ title, onBack }: { title: string; onBack: () => void }) {
  return (
    <View style={styles.subHeader}>
      <TouchableOpacity style={styles.subBackBtn} onPress={onBack} accessibilityRole="button">
        <HugeiconsIcon icon={ArrowLeft01Icon} size={20} color={colors.text} strokeWidth={2} />
      </TouchableOpacity>
      <Text style={styles.subTitle}>{title}</Text>
      <View style={styles.subBackBtn} />
    </View>
  );
}

export function ProfileScreen() {
  const { user, logout } = useAuthStore();
  const queryClient = useQueryClient();
  const [subView, setSubView] = useState<SubView>("main");

  // PIN change state
  const [oldPin, setOldPin] = useState("");
  const [newPin, setNewPin] = useState("");
  const [confirmPin, setConfirmPin] = useState("");
  const [pinLoading, setPinLoading] = useState(false);

  // Language state
  const [langLoading, setLangLoading] = useState(false);

  // KYC state
  const [fan, setFan] = useState("");
  const [fullName, setFullName] = useState("");
  const [dob, setDob] = useState("");
  const [kycLoading, setKycLoading] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["profile"],
    queryFn: () => apiFetch<ProfileResponse>("/identity/me"),
  });

  const p = data?.data;
  const tier = p?.kyc_tier ?? user?.kycTier ?? 0;
  const msisdn = p?.msisdn ?? user?.msisdn ?? "—";
  const currentLang = p?.preferred_lang ?? "am";
  const initials = msisdn.slice(-2).toUpperCase();

  const handleChangePIN = async () => {
    if (newPin.length < 4) { Alert.alert("PIN must be at least 4 digits"); return; }
    if (newPin !== confirmPin) { Alert.alert("PINs do not match"); return; }
    setPinLoading(true);
    try {
      await apiFetch("/identity/pin/change", {
        method: "POST",
        body: JSON.stringify({ old_pin: oldPin, new_pin: newPin }),
      });
      Alert.alert("PIN Changed", "Your PIN has been updated successfully.");
      setOldPin(""); setNewPin(""); setConfirmPin("");
      setSubView("main");
    } catch (e: unknown) {
      Alert.alert("Failed", e instanceof Error ? e.message : "Could not change PIN.");
    } finally {
      setPinLoading(false);
    }
  };

  const handleSetLanguage = async (lang: string) => {
    setLangLoading(true);
    try {
      await apiFetch("/identity/me/language", {
        method: "PATCH",
        body: JSON.stringify({ lang }),
      });
      queryClient.invalidateQueries({ queryKey: ["profile"] });
      Alert.alert("Language Updated", `App language set to ${LANG_LABEL[lang]}.`);
      setSubView("main");
    } catch (e: unknown) {
      Alert.alert("Failed", e instanceof Error ? e.message : "Could not update language.");
    } finally {
      setLangLoading(false);
    }
  };

  const handleKycUpgrade = async () => {
    if (!fan.trim() || !fullName.trim() || !dob.trim()) {
      Alert.alert("Fill in all fields");
      return;
    }
    setKycLoading(true);
    try {
      await apiFetch("/identity/kyc/upgrade", {
        method: "POST",
        body: JSON.stringify({ fan: fan.trim(), full_name: fullName.trim(), dob: dob.trim() }),
      });
      Alert.alert("Submitted", "Your KYC upgrade request has been submitted for branch review.");
      setFan(""); setFullName(""); setDob("");
      setSubView("main");
    } catch (e: unknown) {
      Alert.alert("Failed", e instanceof Error ? e.message : "Could not submit KYC upgrade.");
    } finally {
      setKycLoading(false);
    }
  };

  const handleLogout = () => {
    Alert.alert("Log Out", "Are you sure you want to log out?", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Log Out",
        style: "destructive",
        onPress: async () => {
          try { await apiFetch("/identity/logout", { method: "POST" }); } catch {}
          clearAccessToken();
          logout();
        },
      },
    ]);
  };

  if (subView === "pin") {
    return (
      <SafeAreaView style={styles.safe}>
        <SubHeader title="Change PIN" onBack={() => setSubView("main")} />
        <ScrollView contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">
          <View style={[styles.formCard, shadow.sm]}>
            <Text style={styles.fieldLabel}>Current PIN</Text>
            <TextInput
              style={styles.fieldInput}
              value={oldPin}
              onChangeText={setOldPin}
              keyboardType="number-pad"
              secureTextEntry
              maxLength={6}
              placeholder="••••••"
              placeholderTextColor={colors.textTertiary}
              accessibilityLabel="Current PIN"
            />
            <Text style={[styles.fieldLabel, { marginTop: spacing.md }]}>New PIN</Text>
            <TextInput
              style={styles.fieldInput}
              value={newPin}
              onChangeText={setNewPin}
              keyboardType="number-pad"
              secureTextEntry
              maxLength={6}
              placeholder="••••••"
              placeholderTextColor={colors.textTertiary}
              accessibilityLabel="New PIN"
            />
            <Text style={[styles.fieldLabel, { marginTop: spacing.md }]}>Confirm New PIN</Text>
            <TextInput
              style={styles.fieldInput}
              value={confirmPin}
              onChangeText={setConfirmPin}
              keyboardType="number-pad"
              secureTextEntry
              maxLength={6}
              placeholder="••••••"
              placeholderTextColor={colors.textTertiary}
              accessibilityLabel="Confirm new PIN"
            />
            <TouchableOpacity
              style={[styles.submitBtn, pinLoading && styles.btnDisabled]}
              onPress={handleChangePIN}
              disabled={pinLoading}
              accessibilityRole="button"
            >
              {pinLoading ? <ActivityIndicator color="#FFF" /> : <Text style={styles.submitBtnText}>Change PIN</Text>}
            </TouchableOpacity>
          </View>
        </ScrollView>
      </SafeAreaView>
    );
  }

  if (subView === "language") {
    return (
      <SafeAreaView style={styles.safe}>
        <SubHeader title="App Language" onBack={() => setSubView("main")} />
        <ScrollView contentContainerStyle={styles.content}>
          <View style={[styles.menuCard, shadow.sm]}>
            {LANG_OPTIONS.map((opt, i) => {
              const isSelected = currentLang === opt.key;
              const isLast = i === LANG_OPTIONS.length - 1;
              return (
                <TouchableOpacity
                  key={opt.key}
                  style={[styles.menuItem, !isLast && styles.menuItemBorder]}
                  onPress={() => handleSetLanguage(opt.key)}
                  disabled={langLoading}
                  accessibilityRole="button"
                >
                  <Text style={[styles.menuItemLabel, isSelected && { color: colors.primary, fontWeight: "700" }]}>
                    {opt.label}
                  </Text>
                  {isSelected && (
                    <Text style={{ color: colors.primary, fontWeight: "700" }}>✓</Text>
                  )}
                </TouchableOpacity>
              );
            })}
          </View>
          {langLoading && <ActivityIndicator style={{ marginTop: spacing.lg }} color={colors.primary} />}
        </ScrollView>
      </SafeAreaView>
    );
  }

  if (subView === "kyc") {
    return (
      <SafeAreaView style={styles.safe}>
        <SubHeader title="KYC Upgrade" onBack={() => setSubView("main")} />
        <ScrollView contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">
          <View style={[styles.formCard, shadow.sm]}>
            <Text style={styles.kycNote}>
              Submit your national ID details for branch review to unlock higher transaction limits.
            </Text>
            <Text style={styles.fieldLabel}>Fayda Account Number (FAN)</Text>
            <TextInput
              style={styles.fieldInput}
              value={fan}
              onChangeText={setFan}
              placeholder="Your national ID number"
              placeholderTextColor={colors.textTertiary}
              autoCapitalize="none"
              accessibilityLabel="Fayda Account Number"
            />
            <Text style={[styles.fieldLabel, { marginTop: spacing.md }]}>Full Name</Text>
            <TextInput
              style={styles.fieldInput}
              value={fullName}
              onChangeText={setFullName}
              placeholder="As on your national ID"
              placeholderTextColor={colors.textTertiary}
              accessibilityLabel="Full name"
            />
            <Text style={[styles.fieldLabel, { marginTop: spacing.md }]}>Date of Birth</Text>
            <TextInput
              style={styles.fieldInput}
              value={dob}
              onChangeText={setDob}
              placeholder="YYYY-MM-DD"
              placeholderTextColor={colors.textTertiary}
              accessibilityLabel="Date of birth"
            />
            <TouchableOpacity
              style={[styles.submitBtn, kycLoading && styles.btnDisabled]}
              onPress={handleKycUpgrade}
              disabled={kycLoading}
              accessibilityRole="button"
            >
              {kycLoading ? <ActivityIndicator color="#FFF" /> : <Text style={styles.submitBtnText}>Submit for Review</Text>}
            </TouchableOpacity>
          </View>
        </ScrollView>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.pageHeader}>
          <Text style={styles.pageTitle}>Profile</Text>
        </View>

        <View style={[styles.identityCard, shadow.sm]}>
          <View style={styles.identityLeft}>
            <Text style={styles.identityPhone}>{msisdn}</Text>
            <View style={styles.tierBadge}>
              <View style={[styles.tierDot, { backgroundColor: KYC_TIER_COLOR[tier] ?? colors.warning }]} />
              <Text style={styles.tierText}>{isLoading ? "…" : KYC_TIER_LABEL[tier] ?? "Unknown"}</Text>
            </View>
          </View>
          <View style={[styles.identityAvatar, { backgroundColor: colors.primary }]}>
            <Text style={styles.identityAvatarText}>{initials}</Text>
          </View>
        </View>

        <Text style={styles.sectionLabel}>ACCOUNT</Text>
        <View style={[styles.menuCard, shadow.sm]}>
          <MenuItem icon={UserIcon} label="Your Profile" value={msisdn} />
          <MenuItem
            icon={ShieldUserIcon}
            label="KYC Verification"
            value={KYC_TIER_LABEL[tier]}
            onPress={tier < 2 ? () => setSubView("kyc") : undefined}
          />
          <MenuItem icon={Notification01Icon} label="Notifications" last onPress={() => Alert.alert("Notifications", "Push notifications coming soon.")} />
        </View>

        <Text style={styles.sectionLabel}>FINANCES</Text>
        <View style={[styles.menuCard, shadow.sm]}>
          <MenuItem icon={ChartBarIncreasingIcon} label="Transaction Limits" last value={KYC_TIER_LABEL[tier]} />
        </View>

        <Text style={styles.sectionLabel}>SECURITY</Text>
        <View style={[styles.menuCard, shadow.sm]}>
          <MenuItem icon={LockIcon} label="Change PIN" last onPress={() => setSubView("pin")} />
        </View>

        <Text style={styles.sectionLabel}>OTHERS</Text>
        <View style={[styles.menuCard, shadow.sm]}>
          <MenuItem
            icon={GlobeIcon}
            label="App Language"
            value={LANG_LABEL[currentLang] ?? "Amharic"}
            onPress={() => setSubView("language")}
          />
          <MenuItem icon={HelpCircleIcon} label="Support" last onPress={() => Alert.alert("Support", "Contact us at support@mesobwallet.com")} />
        </View>

        <TouchableOpacity style={styles.logoutBtn} onPress={handleLogout} accessibilityRole="button">
          <Text style={styles.logoutText}>Log Out</Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },
  content: { paddingHorizontal: spacing.md, paddingTop: spacing.md },
  pageHeader: { marginBottom: spacing.lg },
  pageTitle: { ...typography.title },
  identityCard: {
    backgroundColor: colors.surface, borderRadius: radius.lg,
    padding: spacing.lg, flexDirection: "row",
    alignItems: "center", justifyContent: "space-between", marginBottom: spacing.lg,
  },
  identityLeft: { flex: 1 },
  identityPhone: { fontSize: 18, fontWeight: "700", color: colors.text, marginBottom: spacing.sm },
  tierBadge: { flexDirection: "row", alignItems: "center", gap: spacing.xs },
  tierDot: { width: 8, height: 8, borderRadius: 4 },
  tierText: { fontSize: 13, color: colors.textSecondary, fontWeight: "500" },
  identityAvatar: { width: 48, height: 48, borderRadius: 24, alignItems: "center", justifyContent: "center" },
  identityAvatarText: { color: "#FFF", fontSize: 16, fontWeight: "700" },
  sectionLabel: { ...typography.label, marginBottom: spacing.sm, marginTop: spacing.md },
  menuCard: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden", marginBottom: spacing.xs },
  menuItem: {
    flexDirection: "row", alignItems: "center", justifyContent: "space-between",
    paddingHorizontal: spacing.md, paddingVertical: 16,
  },
  menuItemBorder: { borderBottomWidth: 1, borderBottomColor: colors.divider },
  menuItemLeft: { flexDirection: "row", alignItems: "center", flex: 1 },
  menuIconWrap: {
    width: 32, height: 32, borderRadius: 16,
    backgroundColor: colors.primaryLight, alignItems: "center", justifyContent: "center", marginRight: spacing.md,
  },
  menuItemLabel: { fontSize: 15, fontWeight: "500", color: colors.text },
  menuItemDanger: { color: colors.error },
  menuItemRight: {},
  menuItemValue: { fontSize: 13, color: colors.textTertiary },
  logoutBtn: { marginTop: spacing.lg, marginBottom: spacing.sm, alignItems: "center", paddingVertical: spacing.md },
  logoutText: { fontSize: 15, fontWeight: "600", color: colors.error },

  subHeader: {
    flexDirection: "row", alignItems: "center", justifyContent: "space-between",
    paddingHorizontal: spacing.md, paddingVertical: spacing.md,
  },
  subBackBtn: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.surface, alignItems: "center", justifyContent: "center", ...shadow.sm,
  },
  subTitle: { fontSize: 17, fontWeight: "700", color: colors.text },

  formCard: { backgroundColor: colors.surface, borderRadius: radius.lg, padding: spacing.lg },
  fieldLabel: { ...typography.captionBold, color: colors.textSecondary, marginBottom: spacing.sm },
  fieldInput: {
    backgroundColor: colors.bg, borderRadius: radius.md,
    borderWidth: 1.5, borderColor: colors.border,
    paddingHorizontal: spacing.md, height: 52,
    fontSize: 16, color: colors.text,
  },
  submitBtn: {
    backgroundColor: colors.primary, borderRadius: radius.md,
    height: 52, alignItems: "center", justifyContent: "center",
    marginTop: spacing.lg, ...shadow.md,
  },
  btnDisabled: { opacity: 0.55 },
  submitBtnText: { color: "#FFF", fontSize: 16, fontWeight: "700" },
  kycNote: { fontSize: 13, color: colors.textSecondary, marginBottom: spacing.lg, lineHeight: 20 },
});
