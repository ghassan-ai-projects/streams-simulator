import type { ReviewSession } from "../../../app/review";
import { outcomeLine, requestWords } from "../../../model/review";
import { useStore } from "../../primitives";
import styles from "./Interrupts.module.css";

type Reason = "not-offered" | "waiting" | null;

const reasonWords: Readonly<Record<NonNullable<Reason>, string>> = {
  "not-offered": "Interrupt controls are not available: the node has not opened its control socket.",
  waiting: "Waiting for the node to publish its interrupt state…",
};

export function InterruptOutcome({ review, reason }: { readonly review: ReviewSession; readonly reason: Reason }) {
  const { sending, last } = useStore(review.controls);
  const line = last === null ? null : outcomeLine(last.request, last.result, last.check);
  return (
    <>
      {reason !== null && <p className={styles.note}>{reasonWords[reason]}</p>}
      <p className={styles.outcome} role="status" data-tone={sending === null ? line?.tone : "ok"}>
        {sending === null ? (line?.text ?? "") : `${requestWords(sending)}: sent, waiting for the node's answer…`}
      </p>
    </>
  );
}
