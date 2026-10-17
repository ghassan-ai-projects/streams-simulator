import { Info, TriangleAlert } from "lucide-react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { isDirty } from "../../../app/devices";
import { issueSelection, issueTitle } from "../../../model/devices";
import type { FieldIssue } from "../../../model/devices";
import { cx, panel, shared, useCutBottom } from "../../primitives";
import styles from "./Devices.module.css";

interface DetailsProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly issues: readonly FieldIssue[];
}

function Issues({ draft, devices, issues }: DetailsProps) {
  if (issues.length === 0) return null;
  return (
    <section className={panel.group} aria-labelledby="issues-title">
      <h3 id="issues-title">Issues</h3>
      {isDirty(draft) && draft.issues.length > 0 && (
        <p className={panel.hint}>Issues from the binary are from the last save. Save again to check your changes.</p>
      )}
      <ul className={styles.issueList}>
        {issues.map((issue, index) => (
          <li key={`${String(index)}:${issue.message}`}>
            <TriangleAlert size={14} aria-hidden="true" />
            <button
              type="button"
              className={styles.issueButton}
              onClick={() => {
                devices.select(issueSelection(issue));
              }}
            >
              <strong>{issueTitle(draft.doc, issue)}</strong>
              {issue.message}
            </button>
          </li>
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

export function DeviceDetails(props: DetailsProps) {
  const { ref, cut } = useCutBottom<HTMLElement>();
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
        <Issues {...props} />
        <p className={panel.hint}>
          A device is the interface of a board: what it reports and what it accepts. It is saved in the workspace and
          checked by the binary each time you save.
        </p>
        <Technical draft={props.draft} />
      </div>
    </aside>
  );
}
