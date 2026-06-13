import { useState } from "react";
import { View, Text, TouchableOpacity, ScrollView, StyleSheet } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { HugeiconsIcon } from "@hugeicons/react-native";
import { ArrowLeft01Icon, Add01Icon } from "@hugeicons/core-free-icons";
import { IqubGroupList } from "../../../../src/features/iqub/IqubGroupList";
import { IddirGroupList } from "../../../../src/features/iddir/IddirGroupList";
import { IqubGroupDetail } from "../../../../src/features/iqub/IqubGroupDetail";
import { IddirGroupDetail } from "../../../../src/features/iddir/IddirGroupDetail";
import { CreateIqubForm } from "../../../../src/features/iqub/CreateIqubForm";
import { CreateIddirForm } from "../../../../src/features/iddir/CreateIddirForm";
import { JoinGroupForm } from "../../../../src/features/iqub/JoinGroupForm";
import { useAuthStore } from "../../../../src/lib/authStore";
import { colors, spacing, radius, shadow, typography } from "../../../../src/theme/tokens";
import { TAB_BAR_TOTAL_HEIGHT } from "../navigation/AppNavigator";
import type { IqubGroupData } from "../../../../src/features/iqub/IqubGroupDetail";
import type { IddirGroupData } from "../../../../src/features/iddir/IddirGroupDetail";

type CommunityTab = "iqub" | "iddir";

type CommView =
  | { tag: "list" }
  | { tag: "create-iqub" }
  | { tag: "create-iddir" }
  | { tag: "join-iqub" }
  | { tag: "iqub-detail"; group: IqubGroupData }
  | { tag: "iddir-detail"; group: IddirGroupData };

const TABS: { key: CommunityTab; label: string }[] = [
  { key: "iqub", label: "Iqub" },
  { key: "iddir", label: "Iddir" },
];

function viewTitle(view: CommView, tab: CommunityTab): string {
  switch (view.tag) {
    case "create-iqub":  return "Create Iqub Group";
    case "create-iddir": return "Create Iddir Group";
    case "join-iqub":    return "Join Iqub Group";
    case "iqub-detail":  return view.group.name;
    case "iddir-detail": return view.group.name;
    default:             return "Community";
  }
}

export function CommunityScreen() {
  const { user } = useAuthStore();
  const [activeTab, setActiveTab] = useState<CommunityTab>("iqub");
  const [view, setView] = useState<CommView>({ tag: "list" });

  const goList = () => setView({ tag: "list" });

  const isDetail = view.tag === "iqub-detail" || view.tag === "iddir-detail";

  if (view.tag !== "list") {
    return (
      <SafeAreaView style={styles.safe}>
        <View style={styles.subHeader}>
          <TouchableOpacity style={styles.backBtn} onPress={goList} accessibilityRole="button">
            <HugeiconsIcon icon={ArrowLeft01Icon} size={20} color={colors.text} strokeWidth={2} />
          </TouchableOpacity>
          <Text style={styles.subTitle} numberOfLines={1}>{viewTitle(view, activeTab)}</Text>
          <View style={styles.backBtn} />
        </View>

        <ScrollView
          contentContainerStyle={[
            styles.subContent,
            !isDetail && { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 },
            isDetail && { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 },
          ]}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          {view.tag === "create-iqub" && (
            <View style={[styles.card, shadow.sm]}>
              <CreateIqubForm onDone={goList} />
            </View>
          )}
          {view.tag === "create-iddir" && (
            <View style={[styles.card, shadow.sm]}>
              <CreateIddirForm onDone={goList} />
            </View>
          )}
          {view.tag === "join-iqub" && (
            <View style={[styles.card, shadow.sm]}>
              <JoinGroupForm groupId="" onJoined={goList} />
            </View>
          )}
          {view.tag === "iqub-detail" && (
            <IqubGroupDetail
              group={view.group}
              currentUserId={user?.userId ?? ""}
              onBack={goList}
            />
          )}
          {view.tag === "iddir-detail" && (
            <IddirGroupDetail group={view.group} onBack={goList} />
          )}
        </ScrollView>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.pageHeader}>
        <View style={styles.pageHeaderRow}>
          <View>
            <Text style={styles.pageTitle}>Community</Text>
            <Text style={styles.pageSubtitle}>Rotating savings & mutual aid groups</Text>
          </View>
          <View style={styles.headerButtons}>
            {activeTab === "iqub" && (
              <TouchableOpacity
                style={[styles.iconBtn, styles.joinBtn]}
                onPress={() => setView({ tag: "join-iqub" })}
                accessibilityRole="button"
                accessibilityLabel="Join an Iqub group"
              >
                <Text style={styles.joinBtnText}>Join</Text>
              </TouchableOpacity>
            )}
            <TouchableOpacity
              style={[styles.iconBtn, styles.createBtn]}
              onPress={() => setView({ tag: activeTab === "iqub" ? "create-iqub" : "create-iddir" })}
              accessibilityRole="button"
              accessibilityLabel={`Create ${activeTab === "iqub" ? "Iqub" : "Iddir"} group`}
            >
              <HugeiconsIcon icon={Add01Icon} size={20} color="#FFF" strokeWidth={2} />
            </TouchableOpacity>
          </View>
        </View>
      </View>

      <View style={styles.tabRow}>
        {TABS.map((t) => {
          const isActive = activeTab === t.key;
          return (
            <TouchableOpacity
              key={t.key}
              style={[styles.tabPill, isActive && styles.tabPillActive]}
              onPress={() => setActiveTab(t.key)}
              accessibilityRole="tab"
              accessibilityState={{ selected: isActive }}
            >
              <Text style={[styles.tabLabel, isActive && styles.tabLabelActive]}>{t.label}</Text>
            </TouchableOpacity>
          );
        })}
      </View>

      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: TAB_BAR_TOTAL_HEIGHT + 16 }]}
        showsVerticalScrollIndicator={false}
      >
        <View style={[styles.card, shadow.sm]}>
          {activeTab === "iqub" ? (
            <IqubGroupList
              onSelect={(group) => setView({ tag: "iqub-detail", group })}
            />
          ) : (
            <IddirGroupList
              onSelect={(group) => setView({ tag: "iddir-detail", group })}
            />
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bg },

  pageHeader: { paddingHorizontal: spacing.md, paddingTop: spacing.md, marginBottom: spacing.md },
  pageHeaderRow: { flexDirection: "row", alignItems: "center", justifyContent: "space-between" },
  pageTitle: { ...typography.title },
  pageSubtitle: { fontSize: 14, color: colors.textSecondary, marginTop: spacing.xs },

  headerButtons: { flexDirection: "row", gap: spacing.sm, alignItems: "center" },
  iconBtn: {
    height: 40, borderRadius: 20,
    alignItems: "center", justifyContent: "center",
    ...shadow.sm,
  },
  createBtn: {
    width: 40,
    backgroundColor: colors.primary,
  },
  joinBtn: {
    paddingHorizontal: 16,
    backgroundColor: colors.navy,
  },
  joinBtnText: { fontSize: 13, fontWeight: "700", color: "#FFF" },

  tabRow: {
    flexDirection: "row",
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    padding: 4,
    ...shadow.sm,
  },
  tabPill: { flex: 1, paddingVertical: 10, borderRadius: radius.sm, alignItems: "center" },
  tabPillActive: { backgroundColor: colors.primary },
  tabLabel: { fontSize: 14, fontWeight: "600", color: colors.textSecondary },
  tabLabelActive: { color: "#FFF" },

  content: { paddingHorizontal: spacing.md },
  subContent: { paddingHorizontal: spacing.md },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, overflow: "hidden" },

  subHeader: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
  },
  backBtn: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.surface, alignItems: "center", justifyContent: "center",
    shadowColor: "#000", shadowOpacity: 0.06, shadowRadius: 4, elevation: 2,
  },
  subTitle: { fontSize: 17, fontWeight: "700", color: colors.text, flex: 1, textAlign: "center" },
});
