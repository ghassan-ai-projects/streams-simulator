import type { PlantState } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import { instrumentOf } from "../../model/plant";
import { shortChannel } from "../../model/playback";
import { channelWords, showReading } from "../../model/readout";
import { wallClock } from "../../model/report";
import { cx } from "../cx";
import shared from "../shared.module.css";
import { ReportPanel } from "./ReportPanel";
import styles from "./RunScreen.module.css";

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;
type Tab = "report" | "inspect";

interface RunPanelProps {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly unit: (entity: string, state: string) => string;
}

function Readings({
  review,
  state,
  plant,
  entity,
}: {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly entity: string;
}) {
  const readings = review.index.readings(entity, state.positionNs);
  if (readings.length === 0) return <p className={styles.muted}>No observation of {entity} at this time.</p>;
  return (
    <table className={styles.table}>
      <tbody>
        {readings.map((reading) => {
          const shown = showReading(
            reading,
            plant === null ? "" : instrumentOf(plant.plant, entity, shortChannel(reading.channel)),
          );
          return (
            <tr key={reading.channel}>
              <td className={styles.label} title={shown.channel}>
                {channelWords(shown.channel)}
              </td>
              <td className={cx(styles.number, shown.attention !== null && styles[`value-${shown.attention}`])}>
                {shown.value}
                {shown.unit !== "" && <span className={styles.unit}>{shown.unit}</span>}
              </td>
              <td className={styles.clock} title="time of this observation">
                {wallClock(reading.atNs)}
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

function Inspect(props: RunPanelProps) {
  const entity = props.state.selected;
  if (entity === null)
    return <p className={styles.muted}>Select an instance on the graph to see every channel at the playback time.</p>;
  return (
    <>
      <section className={styles.section} aria-label={`Observed · ${entity}`}>
        <h2 className={shared.sectionTitle}>
          {`Observed · ${entity}`}
          <span className={styles.note}>{`at ${wallClock(props.state.positionNs)}`}</span>
        </h2>
        <Readings review={props.review} state={props.state} plant={props.plant} entity={entity} />
      </section>
      {props.state.report.phase === "ready" && (
        <ReportPanel report={props.state.report.value} unit={props.unit} entity={entity} only="kpi" />
      )}
    </>
  );
}

export function RunPanel(props: RunPanelProps) {
  const { state, review } = props;
  const tab = state.tab;
  const tabs: readonly (readonly [Tab, string])[] = [
    ["report", "Report"],
    ["inspect", state.selected === null ? "Inspect" : `Inspect · ${state.selected}`],
  ];
  return (
    <aside className={styles.report} aria-label="Run details" tabIndex={0}>
      <div className={styles.tabs} role="group" aria-label="Run details view">
        {tabs.map(([id, label]) => (
          <button
            key={id}
            type="button"
            aria-pressed={tab === id}
            className={styles.tab}
            onClick={() => {
              review.showTab(id);
            }}
          >
            {label}
          </button>
        ))}
      </div>
      <div role="region" aria-label={tab === "report" ? "Report" : "Inspect"}>
        {tab === "inspect" ? <Inspect {...props} /> : <ReportBody state={state} unit={props.unit} />}
      </div>
    </aside>
  );
}

function ReportBody({
  state,
  unit,
}: {
  readonly state: ReviewState;
  readonly unit: (entity: string, name: string) => string;
}) {
  if (state.report.phase === "ready")
    return <ReportPanel report={state.report.value} unit={unit} entity={null} verify={state.verified} />;
  if (state.report.phase === "loading")
    return (
      <p className={styles.muted} role="status">
        streamsim report is reading the run folder…
      </p>
    );
  return (
    <p className={styles.failure} role="alert">{`${state.report.failure.message} (${state.report.failure.code})`}</p>
  );
}
