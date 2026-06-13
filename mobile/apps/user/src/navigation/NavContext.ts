import { createContext, useContext } from "react";

export type SubScreen = "send-money" | "pay-bills" | "merchant";

export interface NavContextType {
  push: (screen: SubScreen) => void;
  goBack: () => void;
}

export const NavContext = createContext<NavContextType>({
  push: () => {},
  goBack: () => {},
});

export function useNav() {
  return useContext(NavContext);
}
