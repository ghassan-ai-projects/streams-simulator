import { Activity, Monitor, Moon, Sun } from "lucide-react";
import type { ReactNode } from "react";
import type { ThemePreference } from "../../../model/shell";
import { preferenceLabel } from "../../../model/shell";
import labelled from "./Labelled.module.css";
import styles from "./TopBar.module.css";

interface TopBarProps {
  readonly crumbs: (slot: HTMLElement | null) => void;
  readonly status?: ReactNode;
  readonly toggles?: ReactNode;
  readonly help: ReactNode;
  readonly preference: ThemePreference;
  readonly onCycleTheme: () => void;
  readonly onHome: () => void;
}

const themeIcons = { system: Monitor, light: Sun, dark: Moon } as const;
const themeWords = { system: "System", light: "Light", dark: "Dark" } as const;

function ThemeButton({ preference, onCycle }: { readonly preference: ThemePreference; readonly onCycle: () => void }) {
  const ThemeIcon = themeIcons[preference];
  const label = preferenceLabel(preference);
  return (
    <button type="button" className={labelled.labelled} onClick={onCycle} aria-label={label} title={label}>
      <ThemeIcon size={16} strokeWidth={2} aria-hidden="true" />
      {`Theme: ${themeWords[preference]}`}
    </button>
  );
}

export function TopBar({ crumbs, status, toggles, help, preference, onCycleTheme, onHome }: TopBarProps) {
  return (
    <header className={styles.bar}>
      <button
        type="button"
        className={styles.skip}
        onClick={() => {
          document.getElementById("main")?.focus();
        }}
      >
        Skip to the screen
      </button>
      <button type="button" className={styles.brand} onClick={onHome} aria-label="Streams Simulator, workspace home">
        <span className={styles.mark} aria-hidden="true">
          <Activity size={16} strokeWidth={2.25} />
        </span>
        Streams Simulator
      </button>
      <div className={styles.crumbs} ref={crumbs} />
      {status}
      <span className={styles.tools}>
        {toggles}
        {help}
        <ThemeButton preference={preference} onCycle={onCycleTheme} />
      </span>
    </header>
  );
}
