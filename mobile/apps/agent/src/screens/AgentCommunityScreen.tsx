import { useState } from "react";
import {
  View, Text, TouchableOpacity, ScrollView, StyleSheet,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { Add01Icon, ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import { IqubGroupList } from "../../../../src/features/iqub/IqubGroupList";
import { IddirGroupList } from "../../../../src/features/iddir/IddirGroupList";
import { CreateIqubForm } from "../../../../src/features/iqub/CreateIqubForm";
import { CreateIddirForm } from "../../../../src/features/iddir/CreateIddirForm";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";

const PRIMARY = "#16A34A";

type CommunityTab = "iqub" | "iddir";
type Panel = "list" | "create";

const TABS: { key: CommunityTab; label: string }[] = [
  { key: "iqub",  label: "Iqub"  },
  { key: "iddir", label: "Iddir" },
];

export function AgentCommunityScreen() {
  const [activeTab, setActiveTab] = useState<CommunityTab>("iqub");
  const [panel, setPanel] = useState<Panel>("list");

  const handleTabChange = (tab: CommunityTab) => {
    setActiveTab(tab);
    setPanel("list");
  };

  return (
    <SafeAreaView style={styles.safe}>
      {/* Header */}
      <View style={styles.pageHeader}>
        <View>
          <Text style={styles.pageTitle}>Community</Text>
          <Text style={styles.pageSubtitle}>Manage Iqub & Iddir groups</Text>
        </View>

        {panel === "list" ? (
          <TouchableOpacity
            style={styles.createBtn}
            onPress={() => setPanel("create")}
            accessibilityRole="button"
            accessibilityLabel={`Create ${activeTab === "iqub" ? "Iqub" : "Iddir"} group`}
          >
            <HugeiconsIcon icon={Add01Icon} size={16} color="#FFF" strokeWidth={2.2} />
            <Text style={styles.createBtnText}>Create</Text>
          </TouchableOpacity>
        ) : (
          <TouchableOpacity
            style={styles.backBtn}
            onPress={() => setPanel("list")}
            accessibilityRole="button"
          >
            <HugeiconsIcon icon={ArrowLeft01Icon} size={16} color={PRIMARY} strokeWidth={2} />
            <Text style={styles.backBtnText}>Back</Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Tab pills */}
      <View style={[styles.tabRow, shadow.sm]}>
        {TABS.map((t) => {
          const isActive = activeTab === t.key;
          return (
            <TouchableOpacity
              key={t.key}
              style={[styles.tabPill, isActive && styles.tabPillActive]}
              onPress={() => handleTabChange(t.key)}
              accessibilityRole="tab"
              accessibilityState={{ selected: isActive }}
            >
              <Text style={[styles.tabLabel, isActive && styles.tabLabelActive]}>{t.label}</Text>
            </TouchableOpacity>
          );
        })}
      </View>

      {/* Content */}
      <ScrollView
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
      >
        {panel === "list" ? (
          <View style={[styles.card, shadow.sm]}>
            {activeTab === "iqub" ? <IqubGroupList /> : <IddirGroupList />}
          </View>
        ) : (
          <View style={[styles.card, shadow.sm]}>
            {activeTab === "iqub" ? (
              <CreateIqubForm onDone={() => setPanel("list")} />
            ) : (
              <CreateIddirForm onDone={() => setPanel("list")} />
            )}
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },

  pageHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingHorizontal: spacing.md,
    paddingTop: spacing.md,
    marginBottom: spacing.md,
  },
  pageTitle:    { ...typography.heading },
  pageSubtitle: { fontSize: 13, color: colors.textSecondary, marginTop: 2 },

  createBtn: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
    backgroundColor: PRIMARY,
    borderRadius: radius.md,
    paddingHorizontal: spacing.md,
    height: 36,
    ...shadow.sm,
  },
  createBtnText: { color: "#FFF", fontWeight: "700", fontSize: 14 },

  backBtn: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
    borderRadius: radius.md,
    paddingHorizontal: spacing.md,
    height: 36,
  },
  backBtnText: { color: PRIMARY, fontWeight: "700", fontSize: 14 },

  tabRow: {
    flexDirection: "row",
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    padding: 4,
  },
  tabPill: { flex: 1, paddingVertical: 10, borderRadius: radius.sm, alignItems: "center" },
  tabPillActive: { backgroundColor: PRIMARY },
  tabLabel: { fontSize: 14, fontWeight: "600", color: colors.textSecondary },
  tabLabelActive: { color: "#FFF" },

  scrollContent: { paddingHorizontal: spacing.md, paddingBottom: 120 },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },
});
