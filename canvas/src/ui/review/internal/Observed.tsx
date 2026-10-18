import type { ReadyPlant } from "../../../app/plant";
import type { ReviewState, ReviewSession } from "../../../app/review";
import { readyReport, startNsOf } from "../../../app/review";
import { channelOf } from "../../../model/plant";
import { shortChannel } from "../../../model/shared/words";
import { hallOf } from "../../../model/graph";
import type { Shown } from "../../../model/graph";
import { abnormalFirst, channelWords, sharedTimes, showReading, statusWords } from "../../../model/graph";
import { effectorAt, faultedAt, isRealBoard, runClock } from "../../../model/review";
import type { SessionReading } from "../../../model/review";
import { cx } from "../../primitives";
import { OutsideMark } from "../../primitives";
import parts from "./ReportParts.module.css";
import cells from "./ReportTable.module.css";
import styles from "./Observed.module.css";

interface ObservedProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly entity: string;
}

interface ActiveProps {
  readonly faults: readonly string[];
  readonly shown: readonly Shown[];
  readonly effector: string;
}

function ActiveNow({ faults, shown, effector }: ActiveProps) {
  const alarmed = shown.filter((entry) => entry.attention !== null);
  if (faults.length === 0 && alarmed.length === 0 && effector === "") return null;
  return (
    <ul className={styles.activeNow} aria-label="Active now">
      {effector !== "" && <li>{`Effector call: ${effector}`}</li>}
      {faults.map((fault) => (
        <li key={fault} className={styles["value-alarm"]}>{`${statusWords.fault}: ${fault}`}</li>
      ))}
      {alarmed.map((entry) => (
        <li key={entry.channel} className={styles[`value-${entry.attention ?? "warn"}`]}>
          {`${statusWords[entry.attention ?? "warn"]} · ${channelWords(entry.channel)}: ${entry.value}`}
        </li>
      ))}
    </ul>
  );
}

function HallNote({ plant, entity }: { readonly plant: ReadyPlant | null; readonly entity: string }) {
  const members = plant?.plant.instances.filter((instance) => hallOf(instance.id) === entity).length ?? 0;
  return (
    <p className={parts.muted}>
      {members > 0
        ? `Hall ${entity} holds ${String(members)} instances. Open it to see each one's readings.`
        : `No observation of ${entity} at this time.`}
    </p>
  );
}

interface TableProps {
  readonly shown: readonly Shown[];
  readonly at: ReadonlyMap<string, number>;
  readonly startNs: number;
}

function ObservedTable({ shown, at, startNs }: TableProps) {
  const { latest, own } = sharedTimes(shown.map((entry) => at.get(entry.channel) ?? startNs));
  const shared = own.every((time) => time === null);
  return (
    <table className={cells.table}>
      <caption className="sr-only">
        {`${shared ? "Observed at" : "Latest observed at"} ${runClock(latest, startNs)}`}
      </caption>
      <thead className="sr-only">
        <tr>
          <th scope="col">Channel</th>
          <th scope="col">Value</th>
          <th scope="col">Observed at</th>
        </tr>
      </thead>
      <tbody>
        {shown.map((entry, index) => (
          <tr key={entry.channel}>
            <th scope="row" className={cells.label} title={entry.channel}>
              {channelWords(entry.channel)}
            </th>
            <td className={cx(cells.number, entry.attention !== null && styles[`value-${entry.attention}`])}>
              {entry.value}
              {entry.unit !== "" && <span className={styles.unit}>{entry.unit}</span>}
              <OutsideMark text={entry.outside} />
            </td>
            <td className={cells.clock} title="run time of this observation">
              {own[index] === null || own[index] === undefined ? "" : runClock(own[index], startNs)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function BoardReadings({ readings }: { readonly readings: readonly SessionReading[] }) {
  return (
    <table className={cells.table}>
      <caption>Board readings</caption>
      <thead>
        <tr>
          <th scope="col">Reading</th>
          <th scope="col">Value</th>
          <th scope="col">Quality</th>
        </tr>
      </thead>
      <tbody>
        {readings.map((reading) => (
          <tr key={reading.name}>
            <th scope="row">{reading.name}</th>
            <td>{reading.state === "good" ? `${String(reading.value)} ${reading.unit}` : "—"}</td>
            <td>
              {reading.state}
              {reading.quality !== "" && ` · ${reading.quality}`}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function RealBoardNote({ state, entity }: { readonly state: ReviewState; readonly entity: string }) {
  const readings = state.live?.devices.find((entry) => entry.instance === entity)?.readings ?? [];
  return (
    <>
      <p className={parts.muted}>
        {readings.length > 0
          ? "This instance is a real board. These are its own readings, as the board reported them."
          : "This instance is a real board. It has reported no reading at this time."}{" "}
        The simulator&apos;s twin values are never shown as the board&apos;s.
      </p>
      {readings.length > 0 && <BoardReadings readings={readings} />}
    </>
  );
}

export function Observed({ review, state, plant, entity }: ObservedProps) {
  if (isRealBoard(state.origins, entity)) return <RealBoardNote state={state} entity={entity} />;
  const readings = review.index.readings(entity, state.positionNs);
  if (readings.length === 0) return <HallNote plant={plant} entity={entity} />;
  const shown = abnormalFirst(
    readings.map((reading) =>
      showReading(reading, plant === null ? null : channelOf(plant.plant, entity, reading.channel)),
    ),
  );
  const report = readyReport(state);
  const at = new Map(readings.map((reading) => [shortChannel(reading.channel), reading.atNs]));
  return (
    <>
      <ActiveNow
        faults={faultedAt(report, state.positionNs).get(entity) ?? []}
        shown={shown}
        effector={effectorAt(report, state.positionNs).get(entity) ?? ""}
      />
      <ObservedTable shown={shown} at={at} startNs={startNsOf(state, review)} />
    </>
  );
}
