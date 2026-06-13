import { Platform, View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Tabs } from "expo-router";
import { BlurView } from "expo-blur";
import { HugeiconsIcon } from "@hugeicons/react-native";
import type { IconSvgElement } from "@hugeicons/react-native";
import {
  Home01Icon,
  MoneyExchange01Icon,
  UserGroupIcon,
  UserCircleIcon,
} from "@hugeicons/core-free-icons";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { colors, shadow } from "../../../../src/theme/tokens";

const PRIMARY = "#16A34A";
const PRIMARY_LIGHT = "rgba(22,163,74,0.12)";

const iosVersion = Platform.OS === "ios" ? parseFloat(Platform.Version as string) : 0;
const useNativeGlass = iosVersion >= 26;

const TABS: { name: string; label: string; icon: IconSvgElement }[] = [
  { name: "home",       label: "Home",       icon: Home01Icon },
  { name: "operations", label: "Operations", icon: MoneyExchange01Icon },
  { name: "community",  label: "Community",  icon: UserGroupIcon },
  { name: "me",         label: "Me",         icon: UserCircleIcon },
];

const HIDDEN_ROUTES = new Set(["sync-progress"]);

function LiquidTabBar({ state, navigation }: any) {
  const insets = useSafeAreaInsets();
  const bottom = Math.max(insets.bottom, 16) + 8;

  // Hide tab bar on full-screen utility screens
  if (HIDDEN_ROUTES.has(state.routes[state.index]?.name)) return null;

  const inner = (
    <View style={[styles.inner, useNativeGlass && styles.innerNative]}>
      {state.routes.map((route: any, index: number) => {
        const isActive = state.index === index;
        const meta = TABS.find((t) => t.name === route.name) ?? TABS[0];
        return (
          <TouchableOpacity
            key={route.key}
            style={styles.tab}
            onPress={() => navigation.navigate(route.name)}
            accessibilityRole="tab"
            accessibilityLabel={meta.label}
            accessibilityState={{ selected: isActive }}
          >
            {isActive && <View style={styles.pill} />}
            <HugeiconsIcon
              icon={meta.icon}
              size={22}
              color={isActive ? PRIMARY : colors.textTertiary}
              strokeWidth={isActive ? 2.2 : 1.6}
            />
            <Text style={[styles.label, isActive && styles.labelActive]}>{meta.label}</Text>
          </TouchableOpacity>
        );
      })}
    </View>
  );

  const wrapperStyle = [styles.wrapper, shadow.lg, { bottom }];

  if (useNativeGlass) {
    return <View style={[...wrapperStyle, styles.wrapperNative]}>{inner}</View>;
  }

  return (
    <View style={wrapperStyle}>
      <BlurView
        intensity={Platform.OS === "ios" ? 80 : 100}
        tint="systemThickMaterialLight"
        style={styles.blur}
      >
        {inner}
      </BlurView>
    </View>
  );
}

export default function TabLayout() {
  return (
    <Tabs
      tabBar={(props) => <LiquidTabBar {...props} />}
      screenOptions={{ headerShown: false }}
    >
      <Tabs.Screen name="home" />
      <Tabs.Screen name="operations" />
      <Tabs.Screen name="community" />
      <Tabs.Screen name="me" />
      <Tabs.Screen name="sync-progress" options={{ href: null }} />
      <Tabs.Screen name="float"   options={{ href: null }} />
      <Tabs.Screen name="onboard" options={{ href: null }} />
    </Tabs>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    position: "absolute",
    left: 16,
    right: 16,
    height: 64,
    borderRadius: 32,
    overflow: "hidden",
    borderWidth: 0.5,
    borderColor: "rgba(255,255,255,0.7)",
  },
  wrapperNative: {
    backgroundColor: "rgba(242,243,247,0.72)",
    borderColor: "rgba(255,255,255,0.5)",
  },
  blur: { flex: 1, borderRadius: 32, overflow: "hidden" },
  inner: {
    flex: 1,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 8,
    backgroundColor: Platform.OS === "android" ? "rgba(242,243,247,0.95)" : "transparent",
  },
  innerNative: { backgroundColor: "transparent" },
  tab: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    paddingVertical: 10,
    position: "relative",
  },
  pill: {
    position: "absolute",
    top: 6,
    width: 52,
    height: 36,
    borderRadius: 18,
    backgroundColor: PRIMARY_LIGHT,
  },
  label: { fontSize: 10, fontWeight: "500", color: colors.textTertiary, marginTop: 3 },
  labelActive: { color: PRIMARY, fontWeight: "600" },
});
