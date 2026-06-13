import { View, Text, StyleSheet, TouchableOpacity, ActivityIndicator } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";

interface EligibilityResponse {
  data: {
    eligible: boolean;
    tier: string;
    score: number;
    ceiling_minor: number;
    factors: { feature: string; impact: number }[];
  };
}

async function fetchEligibility(): Promise<EligibilityResponse> {
  return apiFetch("/loans/eligibility");
}

interface Props {
  onApply?: () => void;
}

export function LoanEligibilityCard({ onApply }: Props = {}) {
  const { data, isLoading, error } = useQuery({
    queryKey: ["loan-eligibility"],
    queryFn: fetchEligibility,
    retry: false,
  });

  if (isLoading) {
    return (
      <View style={styles.card}>
        <ActivityIndicator color="#1B4FDE" />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.card}>
        <Text style={styles.errorText}>Eligibility check deferred. Try again later.</Text>
      </View>
    );
  }

  const info = data?.data;

  return (
    <View style={styles.card}>
      <Text style={styles.label}>Loan Eligibility</Text>
      {info?.eligible ? (
        <>
          <Text style={styles.amount}>Up to {((info.ceiling_minor ?? 0) / 100).toFixed(0)} ETB</Text>
          <Text style={styles.tier}>Tier {info.tier} · Score {info.score}</Text>
          <TouchableOpacity style={styles.applyBtn} onPress={onApply} accessibilityRole="button">
            <Text style={styles.applyBtnText}>Apply for Loan</Text>
          </TouchableOpacity>
        </>
      ) : (
        <Text style={styles.ineligibleText}>Not eligible yet. Keep saving!</Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    margin: 16,
    padding: 20,
    backgroundColor: "#FFF",
    borderRadius: 16,
    shadowColor: "#000",
    shadowOpacity: 0.06,
    shadowRadius: 6,
    elevation: 3,
  },
  label: { fontSize: 13, color: "#6B7280", marginBottom: 8 },
  amount: { fontSize: 28, fontWeight: "700", color: "#0F1B3D" },
  tier: { fontSize: 13, color: "#6B7280", marginTop: 4, marginBottom: 16 },
  applyBtn: {
    backgroundColor: "#1B4FDE",
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: "center",
  },
  applyBtnText: { color: "#FFF", fontWeight: "600", fontSize: 15 },
  ineligibleText: { color: "#6B7280", fontSize: 14 },
  errorText: { color: "#9CA3AF", fontSize: 13 },
});
