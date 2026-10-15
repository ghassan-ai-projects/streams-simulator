import type { Pins as PortPins } from "../../model/ports";
import type { Shown } from "../../model/readout";
import type { ReviewRow } from "../../model/reviewRows";
import { reviewBody } from "../../model/reviewRows";
import { cx } from "../cx";
import { OutsideMark } from "../OutsideMark";
import styles from "./nodes.module.css";

function Missing({ observed }: { readonly observed: boolean }) {
  const words = observed ? "no reading yet" : "not observed";
  return (
    <span className={styles.notObserved} title={words}>
      <span aria-hidden="true">{observed ? "–" : "·"}</span>
      <span className="sr-only">{words}</span>
    </span>
  );
}

function Value({ row }: { readonly row: ReviewRow }) {
  const { reading } = row;
  if (reading === null) return <Missing observed={row.observed} />;
  return (
    <>
      {reading.value}
      {reading.unit !== "" && <span className={styles.unit}>{reading.unit}</span>}
      <OutsideMark text={reading.outside} />
    </>
  );
}

function ReviewRowItem({ row }: { readonly row: ReviewRow }) {
  const { reading } = row;
  const attention = reading?.attention ?? null;
  return (
    <div
      className={cx(attention !== null && styles[attention], reading === null && !row.observed && styles.unobserved)}
    >
      <dt title={row.label}>{row.label}</dt>
      <dd
        className={reading?.word === true ? styles.word : undefined}
        title={reading?.word === true ? reading.value : undefined}
      >
        <Value row={row} />
      </dd>
    </div>
  );
}

function Note({ fault, effector }: { readonly fault: string; readonly effector: string }) {
  return (
    <div className={styles.note}>
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
  shown,
  observed,
  pins,
  channels,
  fault,
  effector,
}: {
  readonly shown: readonly Shown[];
  readonly observed: ReadonlySet<string>;
  readonly pins: PortPins;
  readonly channels: readonly string[];
  readonly fault: string;
  readonly effector: string;
}) {
  const body = reviewBody(pins, channels, shown, observed);
  return (
    <>
      <Note fault={fault} effector={effector} />
      <dl className={styles.readout}>
        {body.rows.map((row) => (
          <ReviewRowItem key={row.key} row={row} />
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
