import { Info, TriangleAlert } from "lucide-react";
import type { TwinDraft } from "../../../app/devices";
import { twinDomainPath, twinIssueTitle } from "../../../model/devices";
import { cx, panel, shared, useCutBottom } from "../../primitives";
import tw from "./Behaviour.module.css";
import styles from "./Devices.module.css";

function Issues({ draft }: { readonly draft: TwinDraft }) {
  if (draft.issues.length === 0) return null;
  return (
    <section className={panel.group} aria-labelledby="twin-issues-title">
      <h3 id="twin-issues-title">Issues</h3>
      <p className={panel.hint}>These come from the last attempt to create the twin.</p>
      <ul className={styles.issueList}>
        {draft.issues.map((issue, index) => (
          <li key={`${String(index)}:${issue.message}`}>
            <TriangleAlert size={14} aria-hidden="true" />
            <span className={tw.issueText}>
              <strong>{twinIssueTitle(issue)}</strong>
              {issue.message}
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}

export function TwinDetails({ draft }: { readonly draft: TwinDraft }) {
  const { ref, cut } = useCutBottom<HTMLElement>();
  return (
    <aside
      ref={ref}
      className={cx(panel.panel, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-labelledby="twin-details-title"
    >
      <header className={panel.head}>
        <Info size={16} aria-hidden="true" />
        <h2 id="twin-details-title">Twin details</h2>
      </header>
      <div className={panel.body}>
        <Issues draft={draft} />
        <p className={panel.hint}>
          A twin is a plant that behaves like your devices. It is made from the saved devices and your choices, and the
          binary checks that it offers exactly what each device offers.
        </p>
        <details>
          <summary className={cx(shared.sectionTitle, shared.disclosure)}>Technical details</summary>
          <dl className={styles.facts}>
            <dt>Choices file</dt>
            <dd>{draft.choicesPath}</dd>
            <dt>Plant file</dt>
            <dd>{twinDomainPath(draft.doc.id)}</dd>
          </dl>
        </details>
      </div>
    </aside>
  );
}
