import { useEffect } from "react";
import type { ThemeState, ThemeStore } from "../../../app/shell";
import { useStore } from "../../primitives";

export function useTheme(store: ThemeStore): ThemeState {
  const state = useStore(store);
  useEffect(() => {
    document.documentElement.dataset.theme = state.theme;
  }, [state.theme]);
  return state;
}
