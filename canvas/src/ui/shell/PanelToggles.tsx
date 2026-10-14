import { PanelLeft, PanelRight } from "lucide-react";
import { useEffect } from "react";
import type { PanelStore, Side } from "../../app/panelStore";
import { useStore } from "../useStore";
import styles from "./TopBar.module.css";

const keyOf: Readonly<Record<string, Side>> = { "[": "left", "]": "right" };

function typing(target: EventTarget | null): boolean {
  return target instanceof HTMLElement && target.closest("input, textarea, select, [contenteditable='true']") !== null;
}

function usePanelKeys(panels: PanelStore) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const side = keyOf[event.key];
      if (side === undefined || event.ctrlKey || event.metaKey || event.altKey || typing(event.target)) return;
      event.preventDefault();
      panels.toggle(side);
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
    };
  }, [panels]);
}

export function PanelToggles({ panels }: { readonly panels: PanelStore }) {
  const state = useStore(panels);
  usePanelKeys(panels);
  const toggle = (side: Side, label: string, key: string, Icon: typeof PanelLeft) => (
    <button
      type="button"
      className={styles.iconButton}
      aria-pressed={state[side]}
      aria-label={label}
      aria-keyshortcuts={key}
      title={`${label} (${key})`}
      onClick={() => {
        panels.toggle(side);
      }}
    >
      <Icon size={16} aria-hidden="true" />
    </button>
  );
  return (
    <span className={styles.toggles}>
      {toggle("left", "Outline panel", "[", PanelLeft)}
      {toggle("right", "Details panel", "]", PanelRight)}
    </span>
  );
}
