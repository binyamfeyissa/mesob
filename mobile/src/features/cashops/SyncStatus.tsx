import { View, Text, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { useState } from "react";
import { useRouter } from "expo-router";
import { operationQueue, type QueuedOperation } from "../../sync/queue";
import { colors, spacing, radius, shadow, typography } from "../../theme/tokens";

const PRIMARY = "#16A34A";

function formatOp(op: QueuedOperation): string {
  const msisdn = (op.payload as any).customer_msisdn ?? "";
  const amount = (op.payload as any).amount_minor;
  const etb = amount != null ? `${(amount / 100).toFixed(2)} ETB` : "";
  const label = op.type === "CASH_IN" ? "Cash In" : op.type === "CASH_OUT" ? "Cash Out" : "Onboard";
  return [label, msisdn, etb].filter(Boolean).join(" · ");
}

function statusColor(status: QueuedOperation["status"]): string {
  switch (status) {
    case "PENDING":  return colors.warning;
    case "SYNCED":   return colors.success;
    case "REJECTED": return colors.error;
  }
}

export function SyncStatus() {
  const router = useRouter();
  const [ops, setOps] = useState<QueuedOperation[]>(() => operationQueue.getAll());
  const pendingCount = ops.filter((o) => o.status === "PENDING").length;
  const rejectedCount = ops.filter((o) => o.status === "REJECTED").length;

  const refresh = () => setOps([...operationQueue.getAll()]);

  const handleDismiss = (id: string) => {
    operationQueue.dismiss(id);
    refresh();
  };

  const handleRetry = (id: string) => {
    operationQueue.resetToPending(id);
    refresh();
  };

  return (
    <View style={styles.wrapper}>
      {/* Header */}
      <View style={styles.pageHeader}>
        <View>
          <Text style={styles.pageTitle}>Offline Queue</Text>
          <Text style={styles.pageSubtitle}>
            {pendingCount} pending{rejectedCount > 0 ? ` · ${rejectedCount} rejected` : ""}
          </Text>
        </View>
        {pendingCount > 0 && (
          <TouchableOpacity
            style={styles.syncBtn}
            onPress={() => router.push("/(tabs)/sync-progress")}
            accessibilityRole="button"
          >
            <Text style={styles.syncBtnText}>Sync Now</Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Queue list */}
      {ops.length === 0 ? (
        <View style={[styles.emptyCard, shadow.sm]}>
          <Text style={styles.emptyText}>All caught up</Text>
          <Text style={styles.emptySubtext}>No pending or rejected operations.</Text>
        </View>
      ) : (
        <View style={styles.list}>
          {ops.map((item) => {
            const color = statusColor(item.status);
            return (
              <View key={item.id} style={[styles.row, shadow.sm]}>
                <View style={styles.rowLeft}>
                  <Text style={styles.opType}>{formatOp(item)}</Text>
                  <Text style={styles.capturedAt}>
                    {new Date(item.capturedAt).toLocaleString()}
                  </Text>
                  {item.syncError ? (
                    <Text style={styles.errorText} numberOfLines={2}>
                      {item.syncError}
                    </Text>
                  ) : null}
                </View>

                <View style={styles.rowRight}>
                  <View style={[styles.badge, { backgroundColor: color + "20" }]}>
                    <Text style={[styles.badgeText, { color }]}>{item.status}</Text>
                  </View>

                  {item.status === "REJECTED" && (
                    <View style={styles.actionRow}>
                      <TouchableOpacity onPress={() => handleRetry(item.id)} style={styles.actionBtn}>
                        <Text style={[styles.actionText, { color: PRIMARY }]}>Retry</Text>
                      </TouchableOpacity>
                      <TouchableOpacity onPress={() => {
                        Alert.alert(
                          "Dismiss",
                          "Remove this rejected operation from the queue?",
                          [
                            { text: "Cancel", style: "cancel" },
                            { text: "Dismiss", style: "destructive", onPress: () => handleDismiss(item.id) },
                          ]
                        );
                      }} style={styles.actionBtn}>
                        <Text style={[styles.actionText, { color: colors.error }]}>Dismiss</Text>
                      </TouchableOpacity>
                    </View>
                  )}
                </View>
              </View>
            );
          })}
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { paddingHorizontal: spacing.md },

  pageHeader: {
    flexDirection: "row", justifyContent: "space-between",
    alignItems: "center", marginBottom: spacing.md,
  },
  pageTitle: { ...typography.heading },
  pageSubtitle: { fontSize: 13, color: colors.textSecondary, marginTop: 2 },

  syncBtn: {
    backgroundColor: PRIMARY, borderRadius: radius.md,
    paddingHorizontal: spacing.lg, height: 40,
    alignItems: "center", justifyContent: "center", ...shadow.sm,
  },
  syncBtnText: { color: "#FFF", fontWeight: "600", fontSize: 14 },

  emptyCard: {
    backgroundColor: colors.surface, borderRadius: radius.lg,
    padding: spacing.xl, alignItems: "center",
  },
  emptyText: { fontSize: 15, fontWeight: "600", color: colors.text, marginBottom: 4 },
  emptySubtext: { fontSize: 13, color: colors.textTertiary },

  list: { gap: spacing.sm },
  row: {
    flexDirection: "row", justifyContent: "space-between",
    alignItems: "flex-start",
    backgroundColor: colors.surface, borderRadius: radius.md, padding: spacing.md,
  },
  rowLeft: { flex: 1, marginRight: spacing.sm },
  rowRight: { alignItems: "flex-end", gap: 6 },
  opType: { fontSize: 14, fontWeight: "600", color: colors.text },
  capturedAt: { fontSize: 11, color: colors.textTertiary, marginTop: 2 },
  errorText: { fontSize: 12, color: colors.error, marginTop: 4, lineHeight: 16 },

  badge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 4 },
  badgeText: { fontSize: 11, fontWeight: "700" },

  actionRow: { flexDirection: "row", gap: 8 },
  actionBtn: { paddingVertical: 2 },
  actionText: { fontSize: 12, fontWeight: "600" },
});
