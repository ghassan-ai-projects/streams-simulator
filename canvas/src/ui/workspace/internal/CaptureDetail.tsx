import type { Card } from "../../../app/devices";
import {
  captureContent,
  claimedWords,
  collisionWords,
  differenceValues,
  differenceWords,
  unreadableWords,
} from "../../../model/devices";
import type { DraftLines } from "../../../model/devices";
import { cx, shared } from "../../primitives";
import styles from "./Capture.module.css";
import cardStyles from "./CaptureCard.module.css";
import { CaptureIssue } from "./CaptureIssue";

interface IssueRow {
  readonly title: string;
  readonly text: string;
  readonly message: string;
}

function Part({
  title,
  lines,
  none,
}: {
  readonly title: string;
  readonly lines: DraftLines["readings"];
  readonly none: string;
}) {
  return (
    <section className={cardStyles.part}>
      <h4 className={cardStyles.partTitle}>{title}</h4>
      {lines.length === 0 ? (
        <p className={styles.note}>{none}</p>
      ) : (
        <ul className={cardStyles.lines}>
          {lines.map((line, index) => (
            <li key={`${String(index)}:${line.name}`}>
              <strong>{line.name}</strong>
              <span className={cardStyles.reading}>{line.facts}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export function Summary({ lines }: { readonly lines: DraftLines | null }) {
  if (lines === null) return <p className={styles.note}>The announced device could not be read.</p>;
  return (
    <div className={cardStyles.summary}>
      <Part title="Readings" lines={lines.readings} none="This device reports nothing." />
      <Part title="Commands" lines={lines.commands} none="This device accepts no commands." />
    </div>
  );
}

export function Issues({ name, issues }: { readonly name: string; readonly issues: readonly IssueRow[] }) {
  if (issues.length === 0) return null;
  return (
    <div>
      <p className={styles.note}>
        This draft still breaks a rule for device files. It can be saved as a draft and repaired in the editor.
      </p>
      <ul className={styles.issues} aria-label={`Issues in ${name}`}>
        {issues.map((issue, index) => (
          <CaptureIssue key={`${String(index)}:${issue.message}`} {...issue} />
        ))}
      </ul>
    </div>
  );
}

export function Claims({ claims }: { readonly claims: readonly string[] }) {
  const words = claimedWords(claims);
  return words === "" ? null : <p className={cardStyles.claims}>{words}</p>;
}

function savedProblem(card: Card): string {
  return card.saved?.ok === false ? `The saved device could not be read: ${card.saved.failure.diagnostic.message}` : "";
}

export function Comparing({ card }: { readonly card: Card }) {
  const { comparison, draft } = card.kind;
  if (card.decision.kind === "saved" || comparison.status === "new" || comparison.status === "same") return null;
  if (comparison.status === "unreadable")
    return <p className={cardStyles.difference}>{unreadableWords(comparison.file)}</p>;
  const values = card.saved?.ok === true ? differenceValues(draft, card.saved.value.content, comparison.differsAt) : "";
  const problem = savedProblem(card);
  return (
    <div className={cardStyles.difference}>
      <p className={styles.facts}>
        This differs from the saved device {comparison.file}. {differenceWords(draft, comparison.differsAt)} {values}
      </p>
      {comparison.collidesWith !== "" && <p className={styles.facts}>{collisionWords(comparison.collidesWith)}</p>}
      {problem !== "" && (
        <p role="alert" className={styles.problem}>
          {problem}
        </p>
      )}
    </div>
  );
}

export function Fingerprint({ card }: { readonly card: Card }) {
  const { comparison, digest } = card.kind;
  return (
    <details>
      <summary className={cx(shared.sectionTitle, shared.disclosure)}>Fingerprint</summary>
      <dl className={cardStyles.technical}>
        <dt>Digest</dt>
        <dd>{digest}</dd>
        {comparison.status !== "new" && (
          <>
            <dt>Saved file</dt>
            <dd>{comparison.file}</dd>
          </>
        )}
        {(comparison.status === "same" || comparison.status === "differs") && (
          <>
            <dt>Saved digest</dt>
            <dd>{comparison.fileDigest === "" ? "Not reported" : comparison.fileDigest}</dd>
          </>
        )}
        {comparison.status === "differs" && (
          <>
            <dt>First difference at</dt>
            <dd>{comparison.differsAt === "" ? "The whole file" : comparison.differsAt}</dd>
          </>
        )}
      </dl>
    </details>
  );
}

export function RawDraft({ draft }: { readonly draft: Card["kind"]["draft"] }) {
  return (
    <details>
      <summary className={cx(shared.sectionTitle, shared.disclosure)}>Raw draft</summary>
      <pre className={cardStyles.raw}>{captureContent(draft)}</pre>
    </details>
  );
}
