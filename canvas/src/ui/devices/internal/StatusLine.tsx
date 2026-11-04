import type { SaveStatus } from "../../../app/devices";
import styles from "./Devices.module.css";

const tones = {
  busy: styles.statusBusy,
  warn: styles.statusWarn,
  ok: styles.statusOk,
  "": styles.statusIdle,
} as const;

interface StatusLineProps {
  readonly status: SaveStatus;
  readonly id?: string;
}

export function StatusLine({ status, id }: StatusLineProps) {
  return (
    <span id={id} role="status" className={tones[status.tone]} title={status.detail ?? status.text}>
      {status.text}
    </span>
  );
}
