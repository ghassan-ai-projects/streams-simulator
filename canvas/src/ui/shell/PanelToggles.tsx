import { PanelLeft, PanelRight } from "lucide-react";
import { useEffect } from "react";
import type { PanelStore, Side } from "../../app/panelStore";
import { inTextField } from "../keys";
import { useStore } from "../primitives";
import styles from "./TopBar.module.css";

const keyOf: Readonly<Record<string, Side>> = { "[": "left", "]": "right" };

function stillShown(element: Element | null): boolean {
  return (
    element instanceof HTMLElement && element !== document.body && element.checkVisibility({ visibilityProperty: true })
  );
}

function toggleKeepingFocus(panels: PanelStore, side: Side) {
  panels.toggle(side);
  requestAnimationFrame(() => {
    if (stillShown(document.activeElement)) return;
    const graph = document.querySelector<HTMLElement>("[role='application']");
    (graph ?? document.querySelector<HTMLElement>("h1"))?.focus();
  });
}

function usePanelKeys(panels: PanelStore) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const side = keyOf[event.key];
      if (side === undefined || event.ctrlKey || event.metaKey || event.repeat || inTextField(event.target)) return;
      event.preventDefault();
      toggleKeepingFocus(panels, side);
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
    };
  }, [panels]);
}

interface TogglesProps {
  readonly panels: PanelStore;
  readonly right?: string;
  readonly onRightPress?: (() => void) | undefined;
}

export function PanelToggles({ panels, right = "Details", onRightPress }: TogglesProps) {
  const state = useStore(panels);
  usePanelKeys(panels);
  const toggle = (side: Side, word: string, key: string, Icon: typeof PanelLeft, press: () => void) => (
    <button
      type="button"
      className={styles.labelled}
      aria-pressed={state[side]}
      aria-label={`${word} panel`}
      aria-keyshortcuts={key}
      title={`${state[side] ? "Hide" : "Show"} the ${word.toLowerCase()} panel (${key})`}
      onClick={press}
    >
      <Icon size={16} aria-hidden="true" />
      {word}
    </button>
  );
  return (
    <span className={styles.toggles}>
      {toggle("left", "Outline", "[", PanelLeft, () => {
        toggleKeepingFocus(panels, "left");
      })}
      {toggle(
        "right",
        right,
        "]",
        PanelRight,
        onRightPress ??
          (() => {
            toggleKeepingFocus(panels, "right");
          }),
      )}
    </span>
  );
}
