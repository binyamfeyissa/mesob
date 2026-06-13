import { View, StyleSheet } from "react-native";
import { IddirGroupList } from "../../../../src/features/iddir/IddirGroupList";

export function IddirScreen() {
  return (
    <View style={styles.container}>
      <IddirGroupList />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#F5F6FA" },
});
