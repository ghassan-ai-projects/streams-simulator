import type { ActiveScope } from "../../../model/review";
import { cx } from "../../primitives";
import { ModeBadge } from "./ModeBadge";
import styles from "./Interrupts.module.css";

function Chip({ active }: { readonly active: ActiveScope }) {
  return (
    <li
      className={cx(styles.chip, active.latched && styles.latched)}
      data-interrupt={active.latched ? "latched" : active.mode}
    >
      <ModeBadge mode={active.mode} latched={active.latched} size={12} prefix={`${active.label}: `} />
    </li>
  );
}

interface ActiveProps {
  readonly active: readonly ActiveScope[];
  readonly covered: string;
}

export function ActiveInterrupts({ active, covered }: ActiveProps) {
  return (
    <>
      {active.length === 0 ? (
        <p className={styles.note}>No holds or approval modes are set.</p>
      ) : (
        <ul className={styles.chips} aria-label="Active interrupts">
          {active.map((entry) => (
            <Chip key={entry.key} active={entry} />
          ))}
        </ul>
      )}
      {covered !== "" && <p className={styles.note}>{covered}</p>}
    </>
  );
}
