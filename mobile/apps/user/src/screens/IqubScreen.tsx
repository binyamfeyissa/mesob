import { View, StyleSheet } from "react-native";
import { IqubGroupList } from "../../../../src/features/iqub/IqubGroupList";

export function IqubScreen() {
  return (
    <View style={styles.container}>
      <IqubGroupList />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#F5F6FA" },
});
