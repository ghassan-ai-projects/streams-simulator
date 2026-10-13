import { Activity, Monitor, Moon, Sun } from "lucide-react";
import type { ThemePreference } from "../../model/theme";
import { preferenceLabel } from "../../model/theme";
import styles from "./TopBar.module.css";

interface TopBarProps {
  readonly preference: ThemePreference;
  readonly onCycleTheme: () => void;
}

const themeIcons = { system: Monitor, light: Sun, dark: Moon } as const;

export function TopBar({ preference, onCycleTheme }: TopBarProps) {
  const ThemeIcon = themeIcons[preference];
  const label = preferenceLabel(preference);
  return (
    <header className={styles.bar}>
      <div className={styles.brand}>
        <span className={styles.mark} aria-hidden="true">
          <Activity size={16} strokeWidth={2.25} />
        </span>
        Streams Simulator
      </div>
      <div className={styles.spacer} />
      <button type="button" className={styles.iconButton} onClick={onCycleTheme} aria-label={label} title={label}>
        <ThemeIcon size={16} strokeWidth={2} aria-hidden="true" />
      </button>
    </header>
  );
}
