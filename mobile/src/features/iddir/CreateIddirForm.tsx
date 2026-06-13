import { useState } from "react";
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  ActivityIndicator, Alert, ScrollView,
} from "react-native";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

const PRIMARY = "#16A34A";

interface CreateIddirPayload {
  name: string;
  premium_minor: number;
  benefit_minor: number;
  frequency: "MONTHLY" | "QUARTERLY";
}

async function createIddirGroup(payload: CreateIddirPayload): Promise<unknown> {
  return apiFetch("/iddir/groups", { method: "POST", body: JSON.stringify(payload) });
}

const FREQUENCIES: { key: CreateIddirPayload["frequency"]; label: string }[] = [
  { key: "MONTHLY",   label: "Monthly"   },
  { key: "QUARTERLY", label: "Quarterly" },
];

interface Props {
  onDone: () => void;
}

export function CreateIddirForm({ onDone }: Props) {
  const [name, setName]           = useState("");
  const [premiumETB, setPremium]  = useState("");
  const [benefitETB, setBenefit]  = useState("");
  const [frequency, setFrequency] = useState<CreateIddirPayload["frequency"]>("MONTHLY");

  const queryClient = useQueryClient();

  const { mutate, isPending } = useMutation({
    mutationFn: createIddirGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["iddir-groups"] });
      Alert.alert("Iddir Created", `"${name}" is now accepting members.`, [
        { text: "OK", onPress: onDone },
      ]);
    },
    onError: (e: any) => {
      Alert.alert("Failed", e?.message ?? "Could not create Iddir group.");
    },
  });

  const handleSubmit = () => {
    const trimmed    = name.trim();
    const premCents  = Math.round(parseFloat(premiumETB) * 100);
    const benCents   = Math.round(parseFloat(benefitETB) * 100);

    if (!trimmed)                             { Alert.alert("Validation", "Group name is required.");         return; }
    if (isNaN(premCents) || premCents <= 0)   { Alert.alert("Validation", "Enter a valid premium amount.");   return; }
    if (isNaN(benCents)  || benCents  <= 0)   { Alert.alert("Validation", "Enter a valid benefit amount.");   return; }

    mutate({ name: trimmed, premium_minor: premCents, benefit_minor: benCents, frequency });
  };

  return (
    <ScrollView keyboardShouldPersistTaps="handled" showsVerticalScrollIndicator={false}>
      <View style={styles.form}>
        <Text style={styles.formTitle}>Create Iddir Group</Text>

        <View style={styles.field}>
          <Text style={styles.label}>Group Name</Text>
          <TextInput
            style={styles.input}
            value={name}
            onChangeText={setName}
            placeholder="e.g. Kebele Iddir"
            placeholderTextColor={colors.textTertiary}
            returnKeyType="next"
          />
        </View>

        <View style={styles.twoCol}>
          <View style={[styles.field, { flex: 1 }]}>
            <Text style={styles.label}>Premium (ETB)</Text>
            <TextInput
              style={styles.input}
              value={premiumETB}
              onChangeText={setPremium}
              placeholder="e.g. 100"
              placeholderTextColor={colors.textTertiary}
              keyboardType="decimal-pad"
              returnKeyType="next"
            />
          </View>
          <View style={[styles.field, { flex: 1 }]}>
            <Text style={styles.label}>Benefit (ETB)</Text>
            <TextInput
              style={styles.input}
              value={benefitETB}
              onChangeText={setBenefit}
              placeholder="e.g. 5000"
              placeholderTextColor={colors.textTertiary}
              keyboardType="decimal-pad"
            />
          </View>
        </View>

        <View style={styles.field}>
          <Text style={styles.label}>Premium Frequency</Text>
          <View style={styles.segmentRow}>
            {FREQUENCIES.map((f) => {
              const isActive = frequency === f.key;
              return (
                <TouchableOpacity
                  key={f.key}
                  style={[styles.segment, isActive && styles.segmentActive]}
                  onPress={() => setFrequency(f.key)}
                  accessibilityRole="radio"
                  accessibilityState={{ checked: isActive }}
                >
                  <Text style={[styles.segmentLabel, isActive && styles.segmentLabelActive]}>
                    {f.label}
                  </Text>
                </TouchableOpacity>
              );
            })}
          </View>
        </View>

        <TouchableOpacity
          style={[styles.submitBtn, isPending && styles.btnDisabled]}
          onPress={handleSubmit}
          disabled={isPending}
          accessibilityRole="button"
        >
          {isPending ? (
            <ActivityIndicator color="#FFF" />
          ) : (
            <Text style={styles.submitText}>Create Iddir Group</Text>
          )}
        </TouchableOpacity>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  form: { padding: spacing.md, gap: spacing.md },
  formTitle: { fontSize: 16, fontWeight: "700", color: colors.text, marginBottom: spacing.xs },

  field: { gap: 6 },
  label: { fontSize: 13, fontWeight: "600", color: colors.textSecondary },
  input: {
    backgroundColor: colors.bg, borderRadius: radius.md,
    borderWidth: 1, borderColor: colors.border,
    paddingHorizontal: spacing.md, height: 48,
    fontSize: 15, color: colors.text,
  },

  twoCol: { flexDirection: "row", gap: spacing.sm },

  segmentRow: {
    flexDirection: "row",
    backgroundColor: colors.bg,
    borderRadius: radius.md,
    borderWidth: 1, borderColor: colors.border,
    overflow: "hidden",
  },
  segment: { flex: 1, paddingVertical: 11, alignItems: "center" },
  segmentActive: { backgroundColor: PRIMARY },
  segmentLabel: { fontSize: 12, fontWeight: "600", color: colors.textSecondary },
  segmentLabelActive: { color: "#FFF" },

  submitBtn: {
    backgroundColor: PRIMARY, borderRadius: radius.md,
    height: 52, alignItems: "center", justifyContent: "center", ...shadow.sm,
    marginTop: spacing.xs,
  },
  submitText: { fontSize: 15, fontWeight: "700", color: "#FFF" },
  btnDisabled: { opacity: 0.55 },
});
