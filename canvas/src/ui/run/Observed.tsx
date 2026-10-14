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

function ActiveNow({ fault, shown }: { readonly fault: string; readonly shown: readonly Shown[] }) {
  const alarmed = shown.filter((entry) => entry.attention !== null);
  if (fault === "" && alarmed.length === 0) return null;
  return (
    <ul className={styles.activeNow} aria-label="Active now">
      {fault !== "" && <li className={styles["value-alarm"]}>{`${statusWords.fault}: ${fault}`}</li>}
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
      <ActiveNow fault={faultedAt(report, state.positionNs).get(entity) ?? ""} shown={shown} />
      <table className={styles.table}>
        <tbody>
          {shown.map((entry) => (
            <tr key={entry.channel}>
              <td className={styles.label} title={entry.channel}>
                {channelWords(entry.channel)}
              </td>
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
    </>
  );
}
