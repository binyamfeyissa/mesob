import { useState, type ReactElement } from "react";
import { View, StyleSheet } from "react-native";
import { WalletScreen } from "../../src/screens/WalletScreen";
import { SendMoneyScreen } from "../../src/screens/SendMoneyScreen";
import { PayBillsScreen } from "../../src/screens/PayBillsScreen";
import { MerchantScreen } from "../../src/screens/MerchantScreen";
import { NavContext, SubScreen } from "../../src/navigation/NavContext";

export default function HomeTab() {
  const [subScreen, setSubScreen] = useState<SubScreen | null>(null);

  const navCtx = {
    push: (screen: SubScreen) => setSubScreen(screen),
    goBack: () => setSubScreen(null),
  };

  const SubScreenMap: Record<SubScreen, ReactElement> = {
    "send-money": <SendMoneyScreen />,
    "pay-bills": <PayBillsScreen />,
    merchant: <MerchantScreen />,
  };

  return (
    <NavContext.Provider value={navCtx}>
      <View style={styles.root}>
        {subScreen ? SubScreenMap[subScreen] : <WalletScreen />}
      </View>
    </NavContext.Provider>
  );
}

const styles = StyleSheet.create({ root: { flex: 1 } });
