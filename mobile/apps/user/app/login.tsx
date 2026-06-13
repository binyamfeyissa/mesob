import { useRouter } from "expo-router";
import { AuthScreen } from "../src/screens/AuthScreen";

export default function LoginScreen() {
  const router = useRouter();
  return (
    <AuthScreen onAuthenticated={() => router.replace("/(tabs)")} />
  );
}
