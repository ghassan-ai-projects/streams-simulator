import { Info } from "lucide-react";
import type { Held, TwinSession } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { cx, panel, shared } from "../../primitives";
import notes from "./Notes.module.css";

interface HeldProps {
  readonly held: readonly Held[];
  readonly twin: TwinSession;
  readonly navigation: NavigationStore;
}

export function RetryButton({ twin }: Pick<HeldProps, "twin">) {
  return (
    <button
      type="button"
      className={cx(shared.secondary, panel.secondary)}
      onClick={() => {
        void twin.retry();
      }}
    >
      Retry
    </button>
  );
}

export function HeldNotice({ held, twin, navigation }: HeldProps) {
  if (held.length === 0) return null;
  return (
    <section className={notes.held} aria-label="Devices that are not ready for a twin">
      {held.map((entry, index) => (
        <div key={entry.path} className={notes.heldItem}>
          <p className={notes.heldText}>
            <Info size={14} aria-hidden="true" />
            {`${entry.words}.`}
          </p>
          <div className={notes.heldActions}>
            <button
              type="button"
              className={cx(shared.secondary, panel.secondary)}
              onClick={() => {
                navigation.go({ screen: "device", path: entry.path });
              }}
            >
              {`Open ${entry.name}`}
            </button>
            {index === held.length - 1 && <RetryButton twin={twin} />}
          </div>
        </div>
      ))}
    </section>
  );
}
