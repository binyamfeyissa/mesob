import { useEffect, useState, type ReactElement } from "react";
import { View, Text, TouchableOpacity, StyleSheet, ActivityIndicator, Platform } from "react-native";
import { BlurView } from "expo-blur";
import { HugeiconsIcon } from "@hugeicons/react-native";
import type { IconSvgElement } from "@hugeicons/react-native";
import {
  Home01Icon,
  MoneyExchange01Icon,
  UserGroupIcon,
  UserCircleIcon,
} from "@hugeicons/core-free-icons";
import { AgentHomeScreen } from "../screens/AgentHomeScreen";
import { OperationsScreen } from "../screens/OperationsScreen";
import { AgentCommunityScreen } from "../screens/AgentCommunityScreen";
import { ProfileScreen } from "../screens/ProfileScreen";
import { AuthScreen } from "../screens/AuthScreen";
import { useAuthStore } from "../../../../src/lib/authStore";
import { colors, shadow } from "../../../../src/theme/tokens";

type Tab = "home" | "operations" | "community" | "me";

const TABS: { key: Tab; label: string; icon: IconSvgElement }[] = [
  { key: "home",      label: "Home",       icon: Home01Icon },
  { key: "operations", label: "Operations", icon: MoneyExchange01Icon },
  { key: "community",  label: "Community",  icon: UserGroupIcon },
  { key: "me",         label: "Me",         icon: UserCircleIcon },
];

const TAB_BAR_HEIGHT = 64;
const TAB_BAR_BOTTOM = 24;
// Native tab bar handles its own insets; this constant is kept for any remaining screen imports
export const TAB_BAR_TOTAL_HEIGHT = 0;

const AGENT_PRIMARY = "#16A34A";
const AGENT_PRIMARY_LIGHT = "rgba(22,163,74,0.12)";

const iosVersion = Platform.OS === "ios" ? parseFloat(Platform.Version as string) : 0;
const useNativeGlass = iosVersion >= 26;

export function AgentNavigator() {
  const { isReady, isAuthenticated, hydrate } = useAuthStore();
  const [tab, setTab] = useState<Tab>("home");

  useEffect(() => {
    if (!isReady) hydrate();
  }, [isReady, hydrate]);

  if (!isReady) {
    return (
      <View style={styles.splash}>
        <ActivityIndicator color={AGENT_PRIMARY} size="large" />
      </View>
    );
  }

  if (!isAuthenticated) {
    return <AuthScreen onAuthenticated={() => setTab("home")} />;
  }

  const screens: Record<Tab, ReactElement> = {
    home:       <AgentHomeScreen />,
    operations: <OperationsScreen />,
    community:  <AgentCommunityScreen />,
    me:         <ProfileScreen />,
  };

  return (
    <View style={styles.root}>
      <View style={styles.content}>{screens[tab]}</View>
      <TabBar active={tab} onPress={setTab} />
    </View>
  );
}

function TabBar({ active, onPress }: { active: Tab; onPress: (t: Tab) => void }) {
  const inner = (
    <View style={[styles.tabBarInner, useNativeGlass && styles.tabBarInnerNative]}>
      {TABS.map((t) => {
        const isActive = active === t.key;
        return (
          <TouchableOpacity
            key={t.key}
            style={styles.tabItem}
            onPress={() => onPress(t.key)}
            accessibilityRole="tab"
            accessibilityLabel={t.label}
            accessibilityState={{ selected: isActive }}
          >
            {isActive && <View style={styles.activePill} />}
            <HugeiconsIcon
              icon={t.icon}
              size={22}
              color={isActive ? AGENT_PRIMARY : colors.textTertiary}
              strokeWidth={isActive ? 2.2 : 1.6}
            />
            <Text style={[styles.tabLabel, isActive && styles.tabLabelActive]}>{t.label}</Text>
          </TouchableOpacity>
        );
      })}
    </View>
  );

  if (useNativeGlass) {
    return (
      <View style={[styles.tabBarWrapper, styles.tabBarWrapperNative, shadow.lg]}>{inner}</View>
    );
  }

  return (
    <View style={[styles.tabBarWrapper, shadow.lg]}>
      <BlurView
        intensity={Platform.OS === "ios" ? 80 : 100}
        tint="systemThickMaterialLight"
        style={styles.tabBarBlur}
      >
        {inner}
      </BlurView>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.bg },
  splash: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.bg },
  content: { flex: 1 },
  tabBarWrapper: {
    position: "absolute",
    bottom: TAB_BAR_BOTTOM,
    left: 16,
    right: 16,
    height: TAB_BAR_HEIGHT,
    borderRadius: 32,
    overflow: "hidden",
    borderWidth: 0.5,
    borderColor: "rgba(255,255,255,0.7)",
  },
  tabBarWrapperNative: {
    backgroundColor: "rgba(242,243,247,0.72)",
    borderColor: "rgba(255,255,255,0.5)",
  },
  tabBarBlur: { flex: 1, borderRadius: 32, overflow: "hidden" },
  tabBarInner: {
    flex: 1,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 8,
    backgroundColor: Platform.OS === "android" ? "rgba(242,243,247,0.95)" : "transparent",
  },
  tabBarInnerNative: { backgroundColor: "transparent" },
  tabItem: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 10,
    position: "relative",
  },
  activePill: {
    position: "absolute",
    top: 6,
    width: 52,
    height: 36,
    borderRadius: 18,
    backgroundColor: AGENT_PRIMARY_LIGHT,
  },
  tabLabel: { fontSize: 10, fontWeight: "500", color: colors.textTertiary, marginTop: 3 },
  tabLabelActive: { color: AGENT_PRIMARY, fontWeight: "600" },
});
