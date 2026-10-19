import { useRef } from "react";
import type { ControlSession } from "../../../app/review";
import type { NodeNow, PendingApproval } from "../../../model/review";
import { deadlineWords, pendingTitle, waitingWords } from "../../../model/review";
import { shared } from "../../primitives";
import styles from "./Interrupts.module.css";

type Decision = "approve" | "deny";

interface RowProps {
  readonly item: PendingApproval;
  readonly now: NodeNow;
  readonly onDecide: (decision: Decision, commandId: string) => void;
}

function WaitingRow({ item, now, onDecide }: RowProps) {
  return (
    <li className={styles.item}>
      <span className={styles.what} title={pendingTitle(item)}>
        {pendingTitle(item)}
      </span>
      <span className={styles.meta}>{`${item.commandId} · ${deadlineWords(item, now)}`}</span>
      <button
        type="button"
        className={shared.primary}
        aria-label={`Approve ${item.commandId}`}
        onClick={() => {
          onDecide("approve", item.commandId);
        }}
      >
        Approve
      </button>
      <button
        type="button"
        className={shared.secondary}
        aria-label={`Deny ${item.commandId}`}
        onClick={() => {
          onDecide("deny", item.commandId);
        }}
      >
        Deny
      </button>
    </li>
  );
}

interface PendingProps {
  readonly controls: ControlSession;
  readonly pending: readonly PendingApproval[];
  readonly now: NodeNow;
}

export function PendingApprovals({ controls, pending, now }: PendingProps) {
  const heading = useRef<HTMLHeadingElement>(null);
  const decide = (decision: Decision, commandId: string) => {
    const done = decision === "approve" ? controls.approve(commandId) : controls.deny(commandId);
    void done.then(() => {
      heading.current?.focus();
    });
  };
  return (
    <div role="group" aria-label="Commands waiting for approval">
      <h3 className={styles.waitingTitle} tabIndex={-1} ref={heading}>
        {waitingWords(pending.length)}
      </h3>
      {pending.length > 0 && (
        <ul className={styles.list} tabIndex={0} aria-label="Waiting commands">
          {pending.map((item) => (
            <WaitingRow key={item.commandId} item={item} now={now} onDecide={decide} />
          ))}
        </ul>
      )}
    </div>
  );
}
