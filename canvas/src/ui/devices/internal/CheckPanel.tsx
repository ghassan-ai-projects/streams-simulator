import { CircleCheck, CircleHelp, Loader, TriangleAlert } from "lucide-react";
import type { Built, CheckRow, Draft, TwinDraft } from "../../../app/devices";
import { builtOf, checkRows } from "../../../app/devices";
import { partsWords } from "../../../model/devices";
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

function digestOf(built: Built): string {
  const { made, implemented } = built;
  if (made !== null) return made.digest;
  return implemented.kind === "passed" ? implemented.checked.digest : "";
}

function Details({ built }: { readonly built: Built }) {
  const { made } = built;
  const digest = digestOf(built);
  return (
    <details>
      <summary className={cx(shared.sectionTitle, shared.disclosure)}>Technical details</summary>
      <dl className={styles.facts}>
        <dt>Plant file</dt>
        <dd>{built.path}</dd>
        <dt>Digest</dt>
        <dd>{digest === "" ? "Not reported" : digest}</dd>
        {made !== null && (
          <>
            <dt>Parts</dt>
            <dd>{partsWords(made.channels, made.effectors, made.faults)}</dd>
            <dt>Devices</dt>
            <dd>{made.implements.join(", ")}</dd>
          </>
        )}
      </dl>
    </details>
  );
}

export function CheckPanel({ draft, device }: { readonly draft: TwinDraft; readonly device: Draft }) {
  const built = builtOf(draft);
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
      {built !== null && <Details built={built} />}
    </section>
  );
}
