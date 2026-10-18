import { CircleCheck, CircleHelp, Loader, TriangleAlert } from "lucide-react";
import type { CheckFact, CheckRow, Draft, TwinDraft } from "../../../app/devices";
import { checkFacts, checkRows } from "../../../app/devices";
import { cx, panel, shared } from "../../primitives";
import tw from "./Behaviour.module.css";
import styles from "./Devices.module.css";
import { FormTitle } from "./FormTitle";

const icons = {
  ok: <CircleCheck size={16} className={tw.checkOk} aria-hidden="true" />,
  bad: <TriangleAlert size={16} className={tw.checkBad} aria-hidden="true" />,
  pending: <Loader size={16} className={cx(tw.checkIdle, shared.spin)} aria-hidden="true" />,
  idle: <CircleHelp size={16} className={tw.checkIdle} aria-hidden="true" />,
} as const;

function Item({ row }: { readonly row: CheckRow }) {
  return (
    <li className={tw.check}>
      {icons[row.state]}
      <div>
        <span>{row.text}</span>
        {row.notes.length > 0 && (
          <ul className={tw.checkNotes}>
            {row.notes.map((note, index) => (
              <li key={`${String(index)}:${note}`}>{note}</li>
            ))}
          </ul>
        )}
        {row.details.length > 0 && (
          <details>
            <summary>Details</summary>
            {row.details.join(" ")}
          </details>
        )}
      </div>
    </li>
  );
}

function Details({ facts }: { readonly facts: readonly CheckFact[] }) {
  return (
    <details>
      <summary className={cx(shared.sectionTitle, shared.disclosure)}>Technical details</summary>
      <dl className={styles.facts}>
        {facts.map((entry) => (
          <Fact key={entry.label} entry={entry} />
        ))}
      </dl>
    </details>
  );
}

function Fact({ entry }: { readonly entry: CheckFact }) {
  return (
    <>
      <dt>{entry.label}</dt>
      <dd>
        {entry.values.map((value, index) => (
          <span key={`${String(index)}:${value}`} className={tw.factLine}>
            {value}
          </span>
        ))}
      </dd>
    </>
  );
}

export function CheckPanel({ draft, device }: { readonly draft: TwinDraft; readonly device: Draft }) {
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="check-title">
      <FormTitle id="check-title" words="Check" />
      {draft.build.kind === "idle" && (
        <p className={panel.hint}>
          Nothing to check yet. Choose how the twin behaves in the Behaviour step, then create it.
        </p>
      )}
      <ul className={tw.checks}>
        {checkRows(device, draft).map((row) => (
          <Item key={row.id} row={row} />
        ))}
      </ul>
      <Details facts={checkFacts(draft)} />
    </section>
  );
}
