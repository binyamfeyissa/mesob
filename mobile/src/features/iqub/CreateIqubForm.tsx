import { useState } from "react";
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  ActivityIndicator, Alert, ScrollView,
} from "react-native";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

const PRIMARY = "#16A34A";

interface CreateIqubPayload {
  name: string;
  cycle_minor: number;
  frequency: "WEEKLY" | "BIWEEKLY" | "MONTHLY";
  member_limit: number;
}

async function createIqubGroup(payload: CreateIqubPayload): Promise<unknown> {
  return apiFetch("/iqub/groups", { method: "POST", body: JSON.stringify(payload) });
}


const FREQUENCIES: { key: CreateIqubPayload["frequency"]; label: string }[] = [
  { key: "WEEKLY",   label: "Weekly"    },
  { key: "BIWEEKLY", label: "Bi-weekly" },
  { key: "MONTHLY",  label: "Monthly"   },
];

interface Props {
  onDone: () => void;
}

export function CreateIqubForm({ onDone }: Props) {
  const [name, setName] = useState("");
  const [amountETB, setAmountETB] = useState("");
  const [frequency, setFrequency] = useState<CreateIqubPayload["frequency"]>("MONTHLY");
  const [maxMembers, setMaxMembers] = useState("10");

  const queryClient = useQueryClient();

  const { mutate, isPending } = useMutation({
    mutationFn: createIqubGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["iqub-groups"] });
      Alert.alert("Iqub Created", `"${name}" is ready for members to join.`, [
        { text: "OK", onPress: onDone },
      ]);
    },
    onError: (e: any) => {
      Alert.alert("Failed", e?.message ?? "Could not create Iqub group.");
    },
  });

  const handleSubmit = () => {
    const trimmed = name.trim();
    const amountCents = Math.round(parseFloat(amountETB) * 100);
    const members = parseInt(maxMembers, 10);

    if (!trimmed) { Alert.alert("Validation", "Group name is required."); return; }
    if (isNaN(amountCents) || amountCents <= 0) { Alert.alert("Validation", "Enter a valid contribution amount."); return; }
    if (isNaN(members) || members < 2) { Alert.alert("Validation", "At least 2 members required."); return; }

    mutate({ name: trimmed, cycle_minor: amountCents, frequency, member_limit: members });
  };

  return (
    <ScrollView keyboardShouldPersistTaps="handled" showsVerticalScrollIndicator={false}>
      <View style={styles.form}>
        <Text style={styles.formTitle}>Create Iqub Group</Text>

        <View style={styles.field}>
          <Text style={styles.label}>Group Name</Text>
          <TextInput
            style={styles.input}
            value={name}
            onChangeText={setName}
            placeholder="e.g. Addis Savings Circle"
            placeholderTextColor={colors.textTertiary}
            returnKeyType="next"
          />
        </View>

        <View style={styles.field}>
          <Text style={styles.label}>Contribution Amount (ETB)</Text>
          <TextInput
            style={styles.input}
            value={amountETB}
            onChangeText={setAmountETB}
            placeholder="e.g. 500"
            placeholderTextColor={colors.textTertiary}
            keyboardType="decimal-pad"
            returnKeyType="next"
          />
        </View>

        <View style={styles.field}>
          <Text style={styles.label}>Frequency</Text>
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

        <View style={styles.field}>
          <Text style={styles.label}>Max Members</Text>
          <TextInput
            style={styles.input}
            value={maxMembers}
            onChangeText={setMaxMembers}
            placeholder="e.g. 10"
            placeholderTextColor={colors.textTertiary}
            keyboardType="number-pad"
          />
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
            <Text style={styles.submitText}>Create Iqub Group</Text>
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

  segmentRow: {
    flexDirection: "row",
    backgroundColor: colors.bg,
    borderRadius: radius.md,
    borderWidth: 1, borderColor: colors.border,
    overflow: "hidden",
  },
  segment: { flex: 1, paddingVertical: 11, alignItems: "center" },
  segmentActive: { backgroundColor: PRIMARY },
  segmentLabel: { fontSize: 13, fontWeight: "600", color: colors.textSecondary },
  segmentLabelActive: { color: "#FFF" },

  submitBtn: {
    backgroundColor: PRIMARY, borderRadius: radius.md,
    height: 52, alignItems: "center", justifyContent: "center", ...shadow.sm,
    marginTop: spacing.xs,
  },
  submitText: { fontSize: 15, fontWeight: "700", color: "#FFF" },
  btnDisabled: { opacity: 0.55 },
});
