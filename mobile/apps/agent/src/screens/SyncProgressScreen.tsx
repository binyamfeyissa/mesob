import { useEffect, useRef, useState, type ReactElement } from "react";
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Animated,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter } from "expo-router";
import { operationQueue, type QueuedOperation } from "../../../../src/sync/queue";
import { syncPendingOperations } from "../../../../src/sync/engine";
import { getAccessToken } from "../../../../src/api/client";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";

const PRIMARY = "#16A34A";

type ItemOutcome = "PENDING" | "SYNCING" | "APPLIED" | "REJECTED";

interface DisplayOp {
  op: QueuedOperation;
  outcome: ItemOutcome;
  error?: string;
}

function outcomeColor(outcome: ItemOutcome): string {
  switch (outcome) {
    case "APPLIED":  return PRIMARY;
    case "REJECTED": return colors.error;
    case "SYNCING":  return colors.warning;
    default:         return colors.textTertiary;
  }
}

function outcomeLabel(outcome: ItemOutcome): string {
  switch (outcome) {
    case "APPLIED":  return "Applied";
    case "REJECTED": return "Rejected";
    case "SYNCING":  return "Sending…";
    default:         return "Queued";
  }
}

function formatOp(op: QueuedOperation): string {
  const msisdn = (op.payload as any).customer_msisdn ?? "";
  const amount = (op.payload as any).amount_minor;
  const etb = amount != null ? `${(amount / 100).toFixed(2)} ETB` : "";
  const label = op.type === "CASH_IN" ? "Cash In" : op.type === "CASH_OUT" ? "Cash Out" : "Onboard";
  return [label, msisdn, etb].filter(Boolean).join(" · ");
}

function OpRow({ item, anim }: { item: DisplayOp; anim: Animated.Value }): ReactElement {
  const color = outcomeColor(item.outcome);
  return (
    <Animated.View style={[styles.row, shadow.sm, { opacity: anim }]}>
      <View style={styles.rowLeft}>
        <Text style={styles.opLabel}>{formatOp(item.op)}</Text>
        <Text style={styles.capturedAt}>
          {new Date(item.op.capturedAt).toLocaleString()}
        </Text>
        {item.error ? (
          <Text style={styles.errorText} numberOfLines={2}>{item.error}</Text>
        ) : null}
      </View>
      <View style={[styles.badge, { backgroundColor: color + "20" }]}>
        {item.outcome === "SYNCING" ? (
          <ActivityIndicator size="small" color={color} />
        ) : (
          <Text style={[styles.badgeText, { color }]}>{outcomeLabel(item.outcome)}</Text>
        )}
      </View>
    </Animated.View>
  );
}

export function SyncProgressScreen(): ReactElement {
  const router = useRouter();
  const pending = useRef(operationQueue.getPending());

  const [items, setItems] = useState<DisplayOp[]>(
    pending.current.map((op) => ({ op, outcome: "PENDING" as ItemOutcome }))
  );
  const [phase, setPhase] = useState<"idle" | "uploading" | "processing" | "done">("idle");
  const [summary, setSummary] = useState<{ applied: number; rejected: number } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const anims = useRef<Map<string, Animated.Value>>(new Map());

  // Initialize an animation value for each op
  for (const op of pending.current) {
    if (!anims.current.has(op.id)) {
      anims.current.set(op.id, new Animated.Value(0.4));
    }
  }

  function animateIn(id: string) {
    const anim = anims.current.get(id);
    if (anim) Animated.spring(anim, { toValue: 1, useNativeDriver: true, tension: 80, friction: 8 }).start();
  }

  function updateOutcome(id: string, outcome: ItemOutcome, err?: string) {
    setItems((prev) =>
      prev.map((item) =>
        item.op.id === id ? { ...item, outcome, error: err } : item
      )
    );
    animateIn(id);
  }

  useEffect(() => {
    if (pending.current.length === 0) {
      setPhase("done");
      setSummary({ applied: 0, rejected: 0 });
      return;
    }

    let cancelled = false;

    async function run() {
      const token = getAccessToken();
      if (!token) {
        setError("Not signed in — please log in again.");
        return;
      }

      setPhase("uploading");
      // Mark all as SYNCING while uploading
      setItems((prev) => prev.map((item) => ({ ...item, outcome: "SYNCING" })));

      try {
        let appliedCount = 0;
        let rejectedCount = 0;

        const result = await syncPendingOperations(token, (id, outcome, err) => {
          if (cancelled) return;
          setPhase("processing");
          updateOutcome(id, outcome, err);
          if (outcome === "APPLIED") appliedCount++;
          else rejectedCount++;
        });

        if (!cancelled) {
          // Any items still showing SYNCING got no response from server
          setItems((prev) =>
            prev.map((item) =>
              item.outcome === "SYNCING" ? { ...item, outcome: "PENDING" } : item
            )
          );
          setPhase("done");
          setSummary({ applied: result.appliedCount, rejected: result.rejectedCount });
        }
      } catch (e: unknown) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Sync failed — try again.");
          setItems((prev) =>
            prev.map((item) =>
              item.outcome === "SYNCING" ? { ...item, outcome: "PENDING" } : item
            )
          );
          setPhase("done");
        }
      }
    }

    run();
    return () => { cancelled = true; };
  }, []);

  const phaseLabel =
    phase === "uploading"  ? "Uploading operations…" :
    phase === "processing" ? "Processing results…"   :
    phase === "done"       ? "Sync complete"          : "Preparing…";

  return (
    <SafeAreaView style={styles.safe}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backBtn} accessibilityRole="button">
          <Text style={styles.backText}>‹ Back</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Sync Progress</Text>
        <View style={{ width: 60 }} />
      </View>

      {/* Phase banner */}
      <View style={[styles.phaseBanner, phase === "done" && !error && styles.phaseBannerDone, error && styles.phaseBannerError]}>
        {phase !== "done" && <ActivityIndicator size="small" color="#FFF" style={{ marginRight: 8 }} />}
        <Text style={styles.phaseText}>{error ?? phaseLabel}</Text>
      </View>

      {/* Summary chips */}
      {summary && (
        <View style={styles.summaryRow}>
          <View style={[styles.chip, { backgroundColor: PRIMARY + "18" }]}>
            <Text style={[styles.chipText, { color: PRIMARY }]}>
              {summary.applied} applied
            </Text>
          </View>
          {summary.rejected > 0 && (
            <View style={[styles.chip, { backgroundColor: colors.error + "18" }]}>
              <Text style={[styles.chipText, { color: colors.error }]}>
                {summary.rejected} rejected
              </Text>
            </View>
          )}
        </View>
      )}

      {/* Per-op list */}
      <ScrollView
        contentContainerStyle={styles.list}
        showsVerticalScrollIndicator={false}
      >
        {items.length === 0 ? (
          <View style={[styles.emptyCard, shadow.sm]}>
            <Text style={styles.emptyText}>Nothing to sync</Text>
          </View>
        ) : (
          items.map((item) => (
            <OpRow
              key={item.op.id}
              item={item}
              anim={anims.current.get(item.op.id) ?? new Animated.Value(1)}
            />
          ))
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },

  header: {
    flexDirection: "row", alignItems: "center", justifyContent: "space-between",
    paddingHorizontal: spacing.md, paddingTop: spacing.sm, paddingBottom: spacing.sm,
  },
  backBtn: { width: 60 },
  backText: { fontSize: 16, color: PRIMARY, fontWeight: "600" },
  title: { ...typography.heading, fontSize: 17 },

  phaseBanner: {
    flexDirection: "row", alignItems: "center", justifyContent: "center",
    backgroundColor: colors.textTertiary,
    paddingVertical: 10, paddingHorizontal: spacing.md,
    marginHorizontal: spacing.md, borderRadius: radius.md, marginBottom: spacing.md,
  },
  phaseBannerDone:  { backgroundColor: PRIMARY },
  phaseBannerError: { backgroundColor: colors.error },
  phaseText: { color: "#FFF", fontWeight: "600", fontSize: 14 },

  summaryRow: {
    flexDirection: "row", gap: spacing.sm,
    paddingHorizontal: spacing.md, marginBottom: spacing.sm,
  },
  chip: { borderRadius: radius.sm, paddingHorizontal: 12, paddingVertical: 6 },
  chipText: { fontWeight: "700", fontSize: 13 },

  list: { paddingHorizontal: spacing.md, paddingBottom: 40, gap: spacing.sm },

  row: {
    flexDirection: "row", alignItems: "center", justifyContent: "space-between",
    backgroundColor: colors.surface, borderRadius: radius.md, padding: spacing.md,
  },
  rowLeft: { flex: 1, marginRight: spacing.sm },
  opLabel: { fontSize: 14, fontWeight: "600", color: colors.text },
  capturedAt: { fontSize: 11, color: colors.textTertiary, marginTop: 2 },
  errorText: { fontSize: 12, color: colors.error, marginTop: 4, lineHeight: 16 },

  badge: {
    borderRadius: radius.sm, paddingHorizontal: 10, paddingVertical: 6,
    minWidth: 80, alignItems: "center", justifyContent: "center",
  },
  badgeText: { fontSize: 12, fontWeight: "700" },

  emptyCard: {
    backgroundColor: colors.surface, borderRadius: radius.lg,
    padding: spacing.xl, alignItems: "center",
  },
  emptyText: { fontSize: 15, color: colors.textTertiary },
});
