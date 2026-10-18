import type { CaptureState } from "../../../app/devices";
import { refusalLines } from "../../../model/devices";
import { failureLine, shared, useFocusOnMount } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureIssue } from "./CaptureIssue";

type Refused = Extract<CaptureState, { readonly phase: "refused" }>;

interface RefusedProps {
  readonly state: Refused;
  readonly onRetry: () => void;
}

export function CaptureRefused({ state, onRetry }: RefusedProps) {
  const { failure } = state;
  const lines = failure.kind === "refused" ? refusalLines(failure.issues) : [];
  const first = useFocusOnMount<HTMLLIElement>({ when: lines.length > 0 });
  const message = useFocusOnMount<HTMLParagraphElement>({ when: lines.length === 0 });
  return (
    <section className={styles.refusal} aria-labelledby="capture-refused">
      <h2 id="capture-refused" className={styles.panelTitle}>
        The board's announcement was not accepted
      </h2>
      <p ref={message} tabIndex={-1} role="alert" className={styles.refusalMessage}>
        {failure.kind === "refused" ? failure.diagnostic.message : failureLine(failure.diagnostic)}
      </p>
      {lines.length > 0 && (
        <ul className={styles.issues} aria-label="What was wrong">
          {lines.map((line, index) => (
            <CaptureIssue
              key={`${String(index)}:${line.message}`}
              title=""
              text={line.text}
              message={line.message}
              focus={index === 0 ? first : undefined}
            />
          ))}
        </ul>
      )}
      <p className={styles.note}>Nothing was saved.</p>
      <div className={styles.actions}>
        <button type="button" className={shared.primary} onClick={onRetry}>
          Retry
        </button>
      </div>
    </section>
  );
}
