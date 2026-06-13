import { useState } from "react";
import { View, StyleSheet } from "react-native";
import { LoginScreen } from "../../../../src/features/auth/LoginScreen";
import { RegisterScreen } from "../../../../src/features/auth/RegisterScreen";
import { OTPScreen } from "../../../../src/features/auth/OTPScreen";
import { SetPINScreen } from "../../../../src/features/auth/SetPINScreen";

type Step = "login" | "register" | "otp" | "setpin";

interface AuthScreenProps {
  onAuthenticated: () => void;
}

export function AuthScreen({ onAuthenticated }: AuthScreenProps) {
  const [step, setStep] = useState<Step>("login");
  const [registrationId, setRegistrationId] = useState("");
  const [challengeToken, setChallengeToken] = useState("");

  return (
    <View style={styles.container}>
      {step === "login" && (
        <LoginScreen onSuccess={onAuthenticated} onRegister={() => setStep("register")} />
      )}
      {step === "register" && (
        <RegisterScreen onSuccess={(rid) => { setRegistrationId(rid); setStep("otp"); }} />
      )}
      {step === "otp" && (
        <OTPScreen registrationId={registrationId} onSuccess={(ct) => { setChallengeToken(ct); setStep("setpin"); }} />
      )}
      {step === "setpin" && (
        <SetPINScreen challengeToken={challengeToken} onSuccess={onAuthenticated} />
      )}
    </View>
  );
}

const styles = StyleSheet.create({ container: { flex: 1 } });
