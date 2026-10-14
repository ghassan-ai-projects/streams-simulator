import type { PlantState } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import { statesOf } from "../../model/plant";
import { shortChannel } from "../../model/playback";
import { hallOf } from "../../model/plantView";
import type { Shown } from "../../model/readout";
import { abnormalFirst, channelWords, showReading, statusWords } from "../../model/readout";
import { wallClock } from "../../model/report";
import { faultedAt } from "../../model/timeline";
import { cx } from "../cx";
import styles from "./RunScreen.module.css";

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

interface ObservedProps {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly entity: string;
}

function ActiveNow({ faults, shown }: { readonly faults: readonly string[]; readonly shown: readonly Shown[] }) {
  const alarmed = shown.filter((entry) => entry.attention !== null);
  if (faults.length === 0 && alarmed.length === 0) return null;
  return (
    <ul className={styles.activeNow} aria-label="Active now">
      {faults.map((fault) => (
        <li key={fault} className={styles["value-alarm"]}>{`${statusWords.fault}: ${fault}`}</li>
      ))}
      {alarmed.map((entry) => (
        <li key={entry.channel} className={styles[`value-${entry.attention ?? "warn"}`]}>
          {`${statusWords[entry.attention ?? "warn"]}: ${channelWords(entry.channel)} ${entry.value}`}
        </li>
      ))}
    </ul>
  );
}

function HallNote({ plant, entity }: { readonly plant: ReadyPlant | null; readonly entity: string }) {
  const members = plant?.plant.instances.filter((instance) => hallOf(instance.id) === entity).length ?? 0;
  return (
    <p className={styles.muted}>
      {members > 0
        ? `Hall ${entity} holds ${String(members)} instances. Open it to see each one's readings.`
        : `No observation of ${entity} at this time.`}
    </p>
  );
}

function ObservedTable({ shown, at }: { readonly shown: readonly Shown[]; readonly at: ReadonlyMap<string, number> }) {
  return (
    <table className={styles.table}>
      <thead className="sr-only">
        <tr>
          <th scope="col">Channel</th>
          <th scope="col">Value</th>
          <th scope="col">Observed at</th>
        </tr>
      </thead>
      <tbody>
        {shown.map((entry) => (
          <tr key={entry.channel}>
            <th scope="row" className={styles.label} title={entry.channel}>
              {channelWords(entry.channel)}
            </th>
            <td className={cx(styles.number, entry.attention !== null && styles[`value-${entry.attention}`])}>
              {entry.value}
              {entry.unit !== "" && <span className={styles.unit}>{entry.unit}</span>}
            </td>
            <td className={styles.clock} title="time of this observation">
              {wallClock(at.get(entry.channel) ?? 0)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function Observed({ review, state, plant, entity }: ObservedProps) {
  const readings = review.index.readings(entity, state.positionNs);
  if (readings.length === 0) return <HallNote plant={plant} entity={entity} />;
  const shown = abnormalFirst(
    readings.map((reading) =>
      showReading(reading, plant === null ? [] : statesOf(plant.plant, entity, reading.channel)),
    ),
  );
  const report = state.report.phase === "ready" ? state.report.value : null;
  const at = new Map(readings.map((reading) => [shortChannel(reading.channel), reading.atNs]));
  return (
    <>
      <ActiveNow faults={faultedAt(report, state.positionNs).get(entity) ?? []} shown={shown} />
      <ObservedTable shown={shown} at={at} />
    </>
  );
}
