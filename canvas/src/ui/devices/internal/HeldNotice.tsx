import { Info } from "lucide-react";
import type { Held, TwinSession } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { blockingWords, leftOutWords } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import { HeldIssues } from "./HeldIssues";
import notes from "./Notes.module.css";
import type { Reveal } from "./twinProps";

interface HeldProps {
  readonly held: readonly Held[];
  readonly twin: TwinSession;
  readonly navigation: NavigationStore;
  readonly lead: string;
  readonly reveal: Reveal | null;
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

function sayings(entry: Held): string {
  if (entry.issues.length === 0) return `${entry.words}.`;
  return entry.adding ? leftOutWords(entry.name) : blockingWords(entry.name);
}

export function HeldNotice({ held, twin, navigation, lead, reveal }: HeldProps) {
  if (held.length === 0) return null;
  return (
    <section className={notes.held} aria-label="Devices that are not ready for a twin">
      {held.map((entry, index) => (
        <div key={entry.path} className={notes.heldItem}>
          <p className={notes.heldText}>
            <Info size={14} aria-hidden="true" />
            {sayings(entry)}
          </p>
          <HeldIssues issues={entry.issues} reveal={entry.path === lead ? reveal : null} />
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
