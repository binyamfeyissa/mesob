import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ActivityIndicator } from "react-native";
import { apiFetch } from "../../api/client";
import { colors, spacing, radius, shadow } from "../../theme/tokens";

interface JoinGroupFormProps {
  groupId?: string;
  onJoined: () => void;
}

export function JoinGroupForm({ groupId: initialGroupId = "", onJoined }: JoinGroupFormProps) {
  const [groupId, setGroupId] = useState(initialGroupId);
  const [joinCode, setJoinCode] = useState("");
  const [loading, setLoading] = useState(false);

  const handleJoin = async () => {
    const gid = groupId.trim();
    const code = joinCode.trim();
    if (!gid) { Alert.alert("Group ID required", "Paste the Group ID shared by the group leader."); return; }
    if (!code) { Alert.alert("Join code required", "Enter the join code shared by the group leader."); return; }
    setLoading(true);
    try {
      await apiFetch(`/iqub/groups/${gid}/members`, {
        method: "POST",
        body: JSON.stringify({ join_code: code }),
      });
      Alert.alert("Joined!", "You are now a member of this Iqub group.");
      onJoined();
    } catch (e: any) {
      Alert.alert("Join failed", e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.form}>
      <Text style={styles.title}>Join an Iqub Group</Text>
      <Text style={styles.hint}>
        Ask your group leader to share the Group ID and Join Code with you.
      </Text>

      <Text style={styles.label}>Group ID</Text>
      <TextInput
        style={styles.input}
        value={groupId}
        onChangeText={setGroupId}
        autoCapitalize="none"
        autoCorrect={false}
        placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
        placeholderTextColor={colors.textTertiary}
        accessibilityLabel="Iqub group ID"
      />

      <Text style={styles.label}>Join Code</Text>
      <TextInput
        style={[styles.input, styles.codeInput]}
        value={joinCode}
        onChangeText={setJoinCode}
        autoCapitalize="characters"
        autoCorrect={false}
        placeholder="XXXXXXXX"
        placeholderTextColor={colors.textTertiary}
        accessibilityLabel="Iqub group join code"
      />

      <TouchableOpacity
        style={[styles.btn, loading && styles.disabled]}
        onPress={handleJoin}
        disabled={loading}
        accessibilityRole="button"
      >
        {loading
          ? <ActivityIndicator color="#FFF" />
          : <Text style={styles.btnText}>Join Group</Text>
        }
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  form: { padding: spacing.md },
  title: { fontSize: 18, fontWeight: "700", color: colors.text, marginBottom: spacing.xs },
  hint: { fontSize: 13, color: colors.textSecondary, marginBottom: spacing.md, lineHeight: 18 },
  label: { fontSize: 13, fontWeight: "600", color: colors.textSecondary, marginBottom: 6 },
  input: {
    borderWidth: 1, borderColor: colors.border, borderRadius: radius.md,
    paddingHorizontal: spacing.md, paddingVertical: 14,
    fontSize: 14, color: colors.text, backgroundColor: colors.bg, marginBottom: spacing.md,
  },
  codeInput: { fontSize: 20, letterSpacing: 4, textAlign: "center" },
  btn: {
    backgroundColor: colors.primary, borderRadius: radius.md,
    height: 52, alignItems: "center", justifyContent: "center",
    ...shadow.sm, marginTop: spacing.xs,
  },
  disabled: { opacity: 0.5 },
  btnText: { color: "#FFF", fontWeight: "700", fontSize: 16 },
});
