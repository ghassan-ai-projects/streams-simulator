import type { CaptureState } from "../../../app/devices";
import { refusalLines, refusalOutline } from "../../../model/devices";
import { failureLine, shared, useFocusOnMount } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureIssue } from "./CaptureIssue";

type Refused = Extract<CaptureState, { readonly phase: "refused" }>;

interface RefusedProps {
  readonly state: Refused;
  readonly onRetry: () => void;
}

function RawWords({ state }: { readonly state: Refused }) {
  const { failure } = state;
  const lines = failure.kind === "refused" ? refusalLines(failure.issues) : [];
  return (
    <details>
      <summary>Details</summary>
      <p>{failureLine(failure.diagnostic)}</p>
      {lines.map((line, index) => (
        <p key={`${String(index)}:${line.message}`}>{line.message}</p>
      ))}
    </details>
  );
}

export function CaptureRefused({ state, onRetry }: RefusedProps) {
  const { failure } = state;
  const issues = failure.kind === "refused" ? failure.issues : [];
  const outline = refusalOutline(
    failure.kind,
    issues.map((issue) => issue.keyword ?? ""),
  );
  const others = refusalLines(issues).slice(1);
  const sentence = useFocusOnMount<HTMLParagraphElement>();
  return (
    <section className={styles.refusal} aria-labelledby="capture-refused">
      <h2 id="capture-refused" className={styles.panelTitle}>
        {outline.heading}
      </h2>
      <p ref={sentence} tabIndex={-1} role="alert" className={styles.refusalMessage}>
        {outline.sentence} {outline.next}
      </p>
      {others.length > 0 && (
        <ul className={styles.issues} aria-label="Also found">
          {others.map((line, index) => (
            <CaptureIssue key={`${String(index)}:${line.message}`} title="" text={line.text} message={line.message} />
          ))}
        </ul>
      )}
      <p className={styles.note}>Nothing was saved.</p>
      <RawWords state={state} />
      <div className={styles.actions}>
        <button type="button" className={shared.primary} onClick={onRetry}>
          Retry
        </button>
      </div>
    </section>
  );
}
