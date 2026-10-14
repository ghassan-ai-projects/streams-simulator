import { useEffect, useSyncExternalStore } from "react";
import type { ThemeState, ThemeStore } from "../app/themeStore";

export function useTheme(store: ThemeStore): ThemeState {
  const state = useSyncExternalStore(store.subscribe, store.snapshot);
  useEffect(() => {
    document.documentElement.dataset.theme = state.theme;
  }, [state.theme]);
  return state;
}
