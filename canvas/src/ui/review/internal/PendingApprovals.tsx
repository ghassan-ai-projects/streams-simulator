import { useEffect, useRef, useState } from "react";
import type { ControlSession } from "../../../app/review";
import type { NodeNow, PendingApproval } from "../../../model/review";
import { argsWords, arrivalWords, deadlineWords, pendingTitle, waitingWords } from "../../../model/review";
import { shared, useInterval, useStore } from "../../primitives";
import styles from "./Interrupts.module.css";

type Decision = "approve" | "deny";

interface RowProps {
  readonly item: PendingApproval;
  readonly now: NodeNow;
  readonly enabled: boolean;
  readonly onDecide: (decision: Decision, item: PendingApproval) => void;
}

function WaitingRow({ item, now, enabled, onDecide }: RowProps) {
  return (
    <li className={styles.item}>
      <span className={styles.info}>
        <span className={styles.what}>{pendingTitle(item)}</span>
        <span className={styles.meta}>{`${item.commandId} · ${deadlineWords(item, now)}`}</span>
        <span className={styles.meta}>{`Arguments: ${argsWords(item)}`}</span>
      </span>
      <button
        type="button"
        className={shared.primary}
        aria-label={`Approve ${item.commandId}`}
        disabled={!enabled}
        onClick={() => {
          onDecide("approve", item);
        }}
      >
        Approve
      </button>
      <button
        type="button"
        className={shared.secondary}
        aria-label={`Deny ${item.commandId}`}
        disabled={!enabled}
        onClick={() => {
          onDecide("deny", item);
        }}
      >
        Deny
      </button>
    </li>
  );
}

function useArrivals(pending: readonly PendingApproval[]): string {
  const known = useRef<ReadonlySet<string> | null>(null);
  const [said, setSaid] = useState("");
  useEffect(() => {
    const before = known.current;
    known.current = new Set(pending.map((item) => item.commandId));
    const arrived = before === null ? undefined : pending.filter((item) => !before.has(item.commandId)).at(-1);
    if (arrived !== undefined) setSaid(arrivalWords(arrived));
  }, [pending]);
  return said;
}

function useSecondTick(every: number | null): void {
  const [, setTick] = useState(0);
  useInterval(() => {
    setTick((tick) => tick + 1);
  }, every);
}

interface PendingProps {
  readonly controls: ControlSession;
  readonly pending: readonly PendingApproval[];
  readonly world: number;
  readonly now: () => number;
  readonly enabled: boolean;
}

export function PendingApprovals({ controls, pending, world, now, enabled }: PendingProps) {
  const heading = useRef<HTMLHeadingElement>(null);
  const busy = useStore(controls).sending.length > 0;
  const said = useArrivals(pending);
  useSecondTick(pending.length > 0 ? 1000 : null);
  const decide = (decision: Decision, item: PendingApproval) => {
    const done = decision === "approve" ? controls.approve(item) : controls.deny(item);
    void done.then((sent) => {
      if (sent === "sent") heading.current?.focus();
    });
  };
  return (
    <div role="group" aria-label="Commands waiting for approval" className={styles.waiting}>
      <h3 className={styles.waitingTitle} tabIndex={-1} ref={heading}>
        {waitingWords(pending.length)}
      </h3>
      <p className="sr-only" role="status">
        {said}
      </p>
      {pending.length > 0 && (
        <ul className={styles.list} aria-label="Waiting commands">
          {pending.map((item) => (
            <WaitingRow
              key={item.commandId}
              item={item}
              now={{ worldNs: world, wallMs: now() }}
              enabled={enabled && !busy}
              onDecide={decide}
            />
          ))}
        </ul>
      )}
    </div>
  );
}
