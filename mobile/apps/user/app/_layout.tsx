import { useEffect } from "react";
import { ActivityIndicator, View, StyleSheet } from "react-native";
import { Slot, useRouter, useSegments } from "expo-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { useAuthStore } from "../../../src/lib/authStore";

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
});

export default function RootLayout() {
  return (
    <SafeAreaProvider>
      <QueryClientProvider client={queryClient}>
        <AuthGuard />
      </QueryClientProvider>
    </SafeAreaProvider>
  );
}

function AuthGuard() {
  const { isReady, isAuthenticated, hydrate } = useAuthStore();
  const router = useRouter();
  const segments = useSegments();

  useEffect(() => {
    if (!isReady) { hydrate(); return; }
    const inTabsGroup = segments[0] === "(tabs)";
    if (!isAuthenticated && inTabsGroup) {
      router.replace("/login");
    } else if (isAuthenticated && !inTabsGroup && segments[0] !== undefined) {
      router.replace("/(tabs)");
    }
  }, [isAuthenticated, isReady, segments]);

  if (!isReady) {
    return (
      <View style={styles.splash}>
        <ActivityIndicator color="#1B4FDE" size="large" />
      </View>
    );
  }

  return <Slot />;
}

const styles = StyleSheet.create({
  splash: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: "#F5F6FA" },
});
