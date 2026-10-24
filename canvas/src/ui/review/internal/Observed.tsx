import type { ReadyPlant } from "../../../app/plant";
import type { ReviewState, ReviewSession } from "../../../app/review";
import { readyReport } from "../../../app/review";
import { hallOf, statusWords } from "../../../model/graph";
import { deviceOf, effectorAt, faultedAt, isRealBoard, readingWords } from "../../../model/review";
import type { ObservedRow, SessionReading } from "../../../model/review";
import { observedView } from "./ObservedReadings";
import { ObservedTable } from "./ObservedTable";
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
  readonly rows: readonly ObservedRow[];
  readonly effector: string;
}

function ActiveNow({ faults, rows, effector }: ActiveProps) {
  const alarmed = rows.filter((entry) => entry.attention !== null);
  if (faults.length === 0 && alarmed.length === 0 && effector === "") return null;
  return (
    <ul className={styles.activeNow} aria-label="Active now">
      {effector !== "" && <li>{`Effector call: ${effector}`}</li>}
      {faults.map((fault) => (
        <li key={fault} className={styles["value-alarm"]}>{`${statusWords.fault}: ${fault}`}</li>
      ))}
      {alarmed.map((entry) => (
        <li key={entry.key} className={styles[`value-${entry.attention ?? "warn"}`]}>
          {`${statusWords[entry.attention ?? "warn"]} · ${entry.label}: ${entry.value}`}
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
        {readings.map((reading) => {
          const words = readingWords(reading);
          return (
            <tr key={reading.name}>
              <th scope="row">{reading.name}</th>
              <td title={words.unitTitle === "" ? undefined : words.unitTitle}>{words.valueText}</td>
              <td>{words.stateWords}</td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

function RealBoardNote({ state, entity }: { readonly state: ReviewState; readonly entity: string }) {
  const readings = deviceOf(state.live, entity)?.readings ?? [];
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
  const { rows, clock } = observedView(review, state, plant, entity);
  if (rows.length === 0) return <HallNote plant={plant} entity={entity} />;
  const report = readyReport(state);
  return (
    <>
      <ActiveNow
        faults={faultedAt(report, state.positionNs).get(entity) ?? []}
        rows={rows}
        effector={effectorAt(report, state.positionNs).get(entity) ?? ""}
      />
      <ObservedTable rows={rows} clock={clock} />
    </>
  );
}
