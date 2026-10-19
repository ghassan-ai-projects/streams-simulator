import type { ControlState, ReviewSession } from "../../../app/review";
import type { ActionRequest } from "../../../model/review";
import { outcomeLine, requestWords } from "../../../model/review";
import { useStore } from "../../primitives";
import styles from "./Interrupts.module.css";

export type Reason = "not-offered" | "waiting" | "stopping" | "absent-scope";

const reasonWords: Readonly<Record<Reason, string>> = {
  "not-offered": "Interrupt controls are not available: the node has not opened its control socket.",
  waiting: "Waiting for the node to publish its interrupt state…",
  stopping: "The session is stopping, so the node takes no more controls.",
  "absent-scope": "This scope is not in the session now. Pick another scope to hold, release or approve.",
};

interface Shown {
  readonly text: string;
  readonly tone: string | undefined;
}

function sendingLine(controls: ControlState, newest: ActionRequest): Shown {
  const blocked =
    controls.blocked === null ? "" : ` ${outcomeLine(controls.blocked, { kind: "not-sent" }, "none").text}`;
  return { text: `${requestWords(newest)}: sent, waiting for the node's answer…${blocked}`, tone: "ok" };
}

function settledLine(controls: ControlState): Shown {
  if (controls.last === null) return { text: "", tone: undefined };
  const line = outcomeLine(controls.last.request, controls.last.result, controls.last.check);
  const waiting = controls.sending.length > 0 ? " Another control is still waiting for the node." : "";
  return { text: `${line.text}${waiting}`, tone: line.tone };
}

function shownOf(controls: ControlState): Shown {
  const newest = controls.sending.at(-1);
  if (newest === undefined || (controls.last !== null && controls.last.order > newest.order)) {
    return settledLine(controls);
  }
  return sendingLine(controls, newest.request);
}

export function InterruptOutcome({
  review,
  reason,
}: {
  readonly review: ReviewSession;
  readonly reason: Reason | null;
}) {
  const controls = useStore(review.controls);
  const shown = shownOf(controls);
  const busy = controls.sending.length > 0 || controls.last?.check === "pending";
  return (
    <>
      {reason !== null && <p className={styles.note}>{reasonWords[reason]}</p>}
      <p className={styles.outcome} role="status" aria-busy={busy} data-tone={shown.tone}>
        {shown.text}
      </p>
    </>
  );
}
