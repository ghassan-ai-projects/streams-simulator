import { Plus, X } from "lucide-react";
import { othersOf } from "../../../app/devices";
import type { TwinDraft, TwinSession } from "../../../app/devices";
import { kindFacts, stemOf } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import tw from "./Behaviour.module.css";
import styles from "./Devices.module.css";

interface SideProps {
  readonly twin: TwinSession;
  readonly draft: TwinDraft;
}

function Members({ twin, draft }: SideProps) {
  return (
    <ul className={tw.members}>
      {draft.members.map(({ kind }, index) => (
        <li key={kind.kind} className={tw.member}>
          <span className={tw.memberText}>
            <span className={styles.itemName}>{kind.kind}</span>
            <span className={styles.itemFacts}>{kindFacts(kind.metrics.length, kind.commands.length)}</span>
          </span>
          {index > 0 && (
            <button
              type="button"
              className={panel.icon}
              aria-label={`Leave ${kind.kind} out of the twin`}
              onClick={() => {
                twin.leave(kind.kind);
              }}
            >
              <X size={16} aria-hidden="true" />
            </button>
          )}
        </li>
      ))}
    </ul>
  );
}

function Others({ twin, draft }: SideProps) {
  const others = othersOf(draft);
  if (others.length === 0) return <p className={cx(panel.hint, styles.addRow)}>No other devices in the workspace.</p>;
  return (
    <>
      {others.map((entry) => {
        const name = entry.kind || stemOf(entry.path);
        return (
          <button
            key={entry.path}
            type="button"
            className={cx(shared.secondary, tw.addDevice)}
            aria-label={`Add ${name} to the twin`}
            onClick={() => {
              void twin.include(entry.path);
            }}
          >
            <Plus size={14} aria-hidden="true" />
            {name}
          </button>
        );
      })}
    </>
  );
}

export function TwinSide(props: SideProps) {
  return (
    <nav className={styles.outline} aria-label="Devices in the twin">
      <h2 className={shared.sectionTitle}>Devices in the twin</h2>
      <Members {...props} />
      <h2 className={shared.sectionTitle}>Add another device</h2>
      <Others {...props} />
    </nav>
  );
}
