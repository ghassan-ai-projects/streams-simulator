import { Bell } from "lucide-react";
import type { EventMark } from "../../../model/review";
import styles from "./RunScreen.module.css";

interface AlarmJumpProps {
  readonly marks: readonly EventMark[];
  readonly positionNs: number;
  readonly onSeek: (ns: number) => void;
  readonly onEvent: ((mark: EventMark) => void) | undefined;
}

export function AlarmJump({ marks, positionNs, onSeek, onEvent }: AlarmJumpProps) {
  const alarms = marks.filter((mark) => mark.tone === "alarm");
  const alarm = alarms.find((mark) => mark.atNs > positionNs) ?? alarms[0];
  const label = alarm === undefined ? "Jump to alarm: none" : `Jump to alarm: ${alarm.label}`;
  return (
    <button
      type="button"
      className={styles.step}
      aria-label={label}
      title={label}
      aria-disabled={alarm === undefined}
      onClick={() => {
        if (alarm === undefined) return;
        if (onEvent === undefined) onSeek(alarm.atNs);
        else onEvent(alarm);
      }}
    >
      <Bell size={14} aria-hidden="true" />
    </button>
  );
}
