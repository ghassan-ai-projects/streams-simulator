import { Info, TriangleAlert } from "lucide-react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { issueSelection, issueTitle } from "../../../model/devices";
import type { FieldIssue } from "../../../model/devices";
import { cx, panel, shared, useCutBottom } from "../../primitives";
import styles from "./Devices.module.css";
import notes from "./Notes.module.css";

interface DetailsProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly issues: readonly FieldIssue[];
  readonly needs: readonly FieldIssue[];
}

interface ListProps extends Pick<DetailsProps, "draft" | "devices"> {
  readonly id: string;
  readonly title: string;
  readonly items: readonly FieldIssue[];
  readonly twin: boolean;
}

interface ItemProps extends Pick<ListProps, "draft" | "devices" | "twin"> {
  readonly issue: FieldIssue;
}

function IssueItem({ draft, devices, issue, twin }: ItemProps) {
  return (
    <li>
      {twin ? <Info size={14} aria-hidden="true" /> : <TriangleAlert size={14} aria-hidden="true" />}
      <span className={notes.stack}>
        <button
          type="button"
          className={styles.issueButton}
          onClick={() => {
            devices.select(issueSelection(issue));
          }}
        >
          <strong>{issueTitle(draft.doc, issue)}</strong>
          {issue.text}
        </button>
        {issue.message !== issue.text && (
          <details>
            <summary>Details</summary>
            {issue.message}
          </details>
        )}
      </span>
    </li>
  );
}

function IssueList({ id, title, items, twin, draft, devices }: ListProps) {
  if (items.length === 0) return null;
  return (
    <section className={panel.group} aria-labelledby={id} aria-live="polite">
      <h3 id={id}>{title}</h3>
      <ul className={twin ? notes.needList : styles.issueList}>
        {items.map((issue, index) => (
          <IssueItem
            key={`${String(index)}:${issue.message}`}
            issue={issue}
            twin={twin}
            draft={draft}
            devices={devices}
          />
        ))}
      </ul>
    </section>
  );
}

function Technical({ draft }: { readonly draft: Draft }) {
  const digest = draft.check.kind === "valid" ? draft.check.summary.digest : "Not checked yet";
  return (
    <details>
      <summary className={cx(shared.sectionTitle, shared.disclosure)}>Technical details</summary>
      <dl className={styles.facts}>
        <dt>File</dt>
        <dd>{draft.path}</dd>
        <dt>Digest</dt>
        <dd>{digest === "" ? "Not reported" : digest}</dd>
      </dl>
    </details>
  );
}

export function DeviceDetails({ draft, devices, issues, needs }: DetailsProps) {
  const { ref, cut } = useCutBottom<HTMLElement>();
  const list = { draft, devices };
  return (
    <aside
      ref={ref}
      className={cx(panel.panel, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-labelledby="details-title"
    >
      <header className={panel.head}>
        <Info size={16} aria-hidden="true" />
        <h2 id="details-title">Device details</h2>
      </header>
      <div className={panel.body}>
        <IssueList id="issues-title" title="Issues" items={issues} twin={false} {...list} />
        <IssueList id="needs-title" title="Needed for a twin" items={needs} twin {...list} />
        <p className={panel.hint}>
          A device is the interface of a board: what it reports and what it accepts. It is saved in the workspace and
          checked by the binary each time you save.
        </p>
        <Technical draft={draft} />
      </div>
    </aside>
  );
}
