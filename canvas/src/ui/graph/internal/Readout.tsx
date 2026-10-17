import type { ReadoutBody, ReadoutRow } from "../../../model/graph";
import { cx } from "../../primitives";
import { OutsideMark } from "../../primitives";
import styles from "./readout.module.css";

function Missing({ observed }: { readonly observed: boolean }) {
  const words = observed ? "no reading yet" : "not observed";
  return (
    <span className={styles.notObserved} title={words}>
      <span aria-hidden="true">{observed ? "–" : "·"}</span>
      <span className="sr-only">{words}</span>
    </span>
  );
}

function Value({ row }: { readonly row: ReadoutRow }) {
  if (!row.present) return <Missing observed={row.observed} />;
  return (
    <>
      {row.value}
      {row.unit !== "" && <span className={styles.unit}>{row.unit}</span>}
      <OutsideMark text={row.outside} />
    </>
  );
}

function ReadoutRowItem({ row }: { readonly row: ReadoutRow }) {
  const { attention } = row;
  return (
    <div className={cx(attention !== null && styles[attention], !row.present && !row.observed && styles.unobserved)}>
      <dt title={row.label}>{row.label}</dt>
      <dd className={row.word ? styles.word : undefined} title={row.word ? row.value : undefined}>
        <Value row={row} />
      </dd>
    </div>
  );
}

function Note({ fault, effector }: { readonly fault: string; readonly effector: string }) {
  return (
    <div data-noted={fault !== "" ? "true" : undefined} className={cx(styles.note, fault !== "" && styles.noted)}>
      {fault !== "" && <span className={styles.noteFault} title={fault}>{`Fault · ${fault}`}</span>}
      {effector !== "" && (
        <span className={styles.noteCall} title={effector}>
          {effector}
        </span>
      )}
    </div>
  );
}

export function Readout({
  body,
  fault,
  effector,
}: {
  readonly body: ReadoutBody;
  readonly fault: string;
  readonly effector: string;
}) {
  return (
    <>
      <Note fault={fault} effector={effector} />
      <dl className={styles.readout}>
        {body.rows.map((row) => (
          <ReadoutRowItem key={row.key} row={row} />
        ))}
        {body.rows.length === 0 && body.hidden === 0 && <span className={styles.none}>no channels</span>}
        {body.hidden > 0 && (
          <div className={cx(styles.more, body.hiddenAttention !== null && styles[body.hiddenAttention])}>
            <dt>{`+${String(body.hidden)} more`}</dt>
            <dd>in Inspect</dd>
          </div>
        )}
      </dl>
    </>
  );
}
