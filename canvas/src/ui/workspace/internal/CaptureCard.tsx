import { Check, CircleAlert, GitCompare, Plus } from "lucide-react";
import { useId } from "react";
import type { Card, CaptureSession } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { captureLines, comparisonWords, differenceWords, draftIssues, kindFacts } from "../../../model/devices";
import type { DraftLines } from "../../../model/devices";
import { cx, shared } from "../../primitives";
import { CaptureActions } from "./CaptureActions";
import styles from "./Capture.module.css";
import { CaptureIssue } from "./CaptureIssue";

interface CardProps {
  readonly card: Card;
  readonly session: CaptureSession;
  readonly navigation: NavigationStore;
}

type Comparison = Card["kind"]["comparison"];

const looks = {
  new: { Icon: Plus, style: styles.chipNew },
  same: { Icon: Check, style: styles.chipSame },
  differs: { Icon: GitCompare, style: styles.chipDiffers },
} as const;

function Status({ comparison }: { readonly comparison: Comparison }) {
  const { Icon, style } = looks[comparison.status];
  return (
    <span className={cx(styles.chip, style)} title={comparison.status === "new" ? undefined : comparison.file}>
      <Icon size={12} aria-hidden="true" />
      {comparisonWords(comparison)}
    </span>
  );
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
    <section className={styles.part}>
      <h4 className={styles.partTitle}>{title}</h4>
      {lines.length === 0 ? (
        <p className={styles.note}>{none}</p>
      ) : (
        <ul className={styles.lines}>
          {lines.map((line, index) => (
            <li key={`${String(index)}:${line.name}`}>
              <strong>{line.name}</strong>
              <span className={styles.reading}>{line.facts}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function Summary({ lines }: { readonly lines: DraftLines | null }) {
  if (lines === null) return <p className={styles.note}>The announced device could not be read.</p>;
  return (
    <div className={styles.summary}>
      <Part title="Readings" lines={lines.readings} none="This device reports nothing." />
      <Part title="Commands" lines={lines.commands} none="This device accepts no commands." />
    </div>
  );
}

function Issues({ card }: { readonly card: Card }) {
  const issues = draftIssues(card.kind.draft, card.kind.issues);
  if (issues.length === 0) return null;
  return (
    <div>
      <p className={styles.note}>
        This draft still breaks a rule for device files. It can be saved as a draft and repaired in the editor.
      </p>
      <ul className={styles.issues} aria-label={`Issues in ${card.kind.kind}`}>
        {issues.map((issue, index) => (
          <CaptureIssue key={`${String(index)}:${issue.message}`} {...issue} />
        ))}
      </ul>
    </div>
  );
}

function Difference({ card }: { readonly card: Card }) {
  const { comparison, draft } = card.kind;
  if (comparison.status !== "differs") return null;
  return (
    <p className={styles.difference}>
      This differs from the saved device {comparison.file}. {differenceWords(draft, comparison.differsAt)}
    </p>
  );
}

function Technical({ comparison, digest }: { readonly comparison: Comparison; readonly digest: string }) {
  return (
    <details>
      <summary className={cx(shared.sectionTitle, shared.disclosure)}>Details</summary>
      <dl className={styles.technical}>
        <dt>Digest</dt>
        <dd>{digest}</dd>
        {comparison.status !== "new" && (
          <>
            <dt>Saved file</dt>
            <dd>{comparison.file}</dd>
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

function Head({ card, id, lines }: { readonly card: Card; readonly id: string; readonly lines: DraftLines | null }) {
  const facts = lines === null ? kindFacts(null, null) : kindFacts(lines.readings.length, lines.commands.length);
  return (
    <header className={styles.cardHead}>
      <h3 id={id} className={styles.cardName}>
        {card.kind.kind}
      </h3>
      <span className={styles.facts}>{facts}</span>
      <Status comparison={card.kind.comparison} />
      {card.kind.issues.length > 0 && (
        <span className={cx(styles.chip, styles.chipInvalid)}>
          <CircleAlert size={12} aria-hidden="true" />
          Not yet valid
        </span>
      )}
    </header>
  );
}

export function CaptureCard(props: CardProps) {
  const id = useId();
  const { kind } = props.card;
  const lines = captureLines(kind.draft);
  return (
    <li>
      <article className={styles.card} aria-labelledby={id}>
        <Head card={props.card} id={id} lines={lines} />
        <Difference card={props.card} />
        <Issues card={props.card} />
        <Summary lines={lines} />
        <Technical comparison={kind.comparison} digest={kind.digest} />
        <CaptureActions {...props} />
      </article>
    </li>
  );
}
