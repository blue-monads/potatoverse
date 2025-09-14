"use client"
import { createContext, useContext, ReactNode } from "react";
import { useGAppState, GAppStateHandle } from "@/hooks";

const AppContext = createContext<GAppStateHandle | null>(null);

export const useGApp = () => {
  const context = useContext(AppContext);
  if (!context) {
    throw new Error("useGAppState must be used within an GAppStateContext");
  }
  return context;
};

interface GAppStateContextProps {
  children: ReactNode;
}

export const GAppStateContext = ({ children }: GAppStateContextProps) => {
  const appState = useGAppState();

  return (
      <AppContext.Provider value={appState}>
      {children}
    </AppContext.Provider>
  );
};
