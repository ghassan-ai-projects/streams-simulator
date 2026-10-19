import type { ReactNode } from "react";
import { ClipboardCheck, Hand, OctagonX } from "lucide-react";
import type { ActiveScope } from "../../../model/review";
import { cx } from "../../primitives";
import styles from "./Interrupts.module.css";

function iconOf(active: ActiveScope): ReactNode {
  if (active.latched) return <OctagonX size={12} aria-hidden="true" />;
  return active.mode === "hold" ? (
    <Hand size={12} aria-hidden="true" />
  ) : (
    <ClipboardCheck size={12} aria-hidden="true" />
  );
}

function Chip({ active }: { readonly active: ActiveScope }) {
  return (
    <li
      className={cx(styles.chip, active.latched && styles.latched)}
      data-interrupt={active.latched ? "latched" : active.mode}
    >
      {iconOf(active)}
      {active.text}
    </li>
  );
}

export function ActiveInterrupts({
  active,
  noteId,
}: {
  readonly active: readonly ActiveScope[];
  readonly noteId: string;
}) {
  return (
    <div className={styles.row}>
      {active.length === 0 ? (
        <p className={styles.note}>No holds or approval modes are set.</p>
      ) : (
        <ul className={styles.chips} aria-label="Active interrupts">
          {active.map((entry) => (
            <Chip key={entry.key} active={entry} />
          ))}
        </ul>
      )}
      <p className={styles.note} id={noteId}>
        Emergency stop sends each device in scope its safe stop and holds the scope, latched, until you release it.
      </p>
    </div>
  );
}
