import { Pin } from "lucide-react";
import type { ObservedRow } from "../../../model/review";
import styles from "./ObservedTable.module.css";

export interface WatchControls {
  readonly watched: (channel: string) => boolean;
  readonly full: boolean;
  readonly toggle: (channel: string) => void;
}

export function WatchButton({ row, watch }: { readonly row: ObservedRow; readonly watch: WatchControls }) {
  const on = watch.watched(row.key);
  const blocked = watch.full && !on;
  return (
    <button
      type="button"
      className={styles.watch}
      aria-pressed={on}
      aria-label={`Watch ${row.label}`}
      title={blocked ? "The watch list is full" : on ? "Stop watching this channel" : "Watch this channel in Trends"}
      disabled={blocked}
      onClick={() => {
        watch.toggle(row.key);
      }}
    >
      <Pin size={13} aria-hidden="true" />
    </button>
  );
}
