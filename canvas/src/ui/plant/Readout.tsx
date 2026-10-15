import { Handle, Position } from "@xyflow/react";
import type { Pin, Pins as PortPins } from "../../model/ports";
import type { Shown } from "../../model/readout";
import { channelWords, displayedRows } from "../../model/readout";
import { pinHandle } from "../../model/plantView";
import { cx } from "../cx";
import { OutsideMark } from "../OutsideMark";
import styles from "./nodes.module.css";

function rowHandles(side: "in" | "out", pin: Pin) {
  const type = side === "in" ? "target" : "source";
  const near = side === "in" ? Position.Left : Position.Right;
  const far = side === "in" ? Position.Right : Position.Left;
  return [
    <Handle
      key={pin.name}
      type={type}
      position={near}
      id={pin.name}
      className={cx(styles.rowAnchor, near === Position.Left ? styles.rowLeft : styles.rowRight)}
      isConnectable={false}
    />,
    <Handle
      key={pinHandle(pin.name, side, true)}
      type={type}
      position={far}
      id={pinHandle(pin.name, side, true)}
      className={cx(styles.rowAnchor, far === Position.Left ? styles.rowLeft : styles.rowRight)}
      isConnectable={false}
    />,
  ];
}

function RowPins({ channel, pins }: { readonly channel: string; readonly pins: PortPins }) {
  return (
    <>
      {pins.ins.filter((pin) => pin.channel === channel).flatMap((pin) => rowHandles("in", pin))}
      {pins.outs.filter((pin) => pin.channel === channel).flatMap((pin) => rowHandles("out", pin))}
    </>
  );
}

function ReadoutRow({
  entry,
  pins,
  wired,
}: {
  readonly entry: Shown;
  readonly pins: PortPins;
  readonly wired: boolean;
}) {
  return (
    <div className={entry.attention === null ? undefined : styles[entry.attention]}>
      <dt title={entry.channel}>
        {channelWords(entry.channel)}
        {wired && <RowPins channel={entry.channel} pins={pins} />}
      </dt>
      <dd className={entry.word ? styles.word : undefined} title={entry.word ? entry.value : undefined}>
        {entry.value}
        {entry.unit !== "" && <span className={styles.unit}>{entry.unit}</span>}
        <OutsideMark text={entry.outside} />
      </dd>
    </div>
  );
}

export function Readout({
  shown,
  pins,
  limit,
}: {
  readonly shown: readonly Shown[];
  readonly pins: PortPins;
  readonly limit: number;
}) {
  if (shown.length === 0) {
    return (
      <div className={styles.readout}>
        <span className={styles.none}>no observation yet</span>
      </div>
    );
  }
  const more = shown.length - 4;
  return (
    <dl className={styles.readout}>
      {displayedRows(shown, Number.POSITIVE_INFINITY).map((entry, index) => (
        <ReadoutRow key={entry.channel} entry={entry} pins={pins} wired={index < limit} />
      ))}
      {more > 0 && (
        <div className={styles.more}>
          <dt>{`+${String(more + 1)} more`}</dt>
          <dd>in Inspect</dd>
        </div>
      )}
    </dl>
  );
}
