import { useEffect } from "react";
import { registerRootComponent } from "expo";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { AgentNavigator } from "./navigation/AgentNavigator";
import { operationQueue } from "../../../src/sync/queue";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
});

function App() {
  useEffect(() => {
    operationQueue.hydrate();
  }, []);

  return (
    <SafeAreaProvider>
      <QueryClientProvider client={queryClient}>
        <AgentNavigator />
      </QueryClientProvider>
    </SafeAreaProvider>
  );
}

registerRootComponent(App);
