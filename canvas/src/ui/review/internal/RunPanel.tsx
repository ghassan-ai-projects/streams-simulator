import { failureLine, shared } from "../../primitives";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewState, ReviewSession } from "../../../app/review";
import { startNsOf } from "../../../app/review";
import { originWords } from "../../../model/graph";
import { runClock } from "../../../model/review";
import type { InstanceOrigin } from "../../../model/review";
import { OriginMark } from "../../graph";
import { cx } from "../../primitives";
import { useCutBottom } from "../../primitives";
import { Observed } from "./Observed";
import { ReportPanel } from "./ReportPanel";
import parts from "./ReportParts.module.css";
import styles from "./RunPanel.module.css";

type Tab = "report" | "inspect";

interface RunPanelProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly unit: (entity: string, state: string) => string;
}

function startOf({ state, review }: RunPanelProps): number {
  return startNsOf(state, review);
}

function OriginLine({ origin }: { readonly origin: InstanceOrigin | undefined }) {
  if (origin === undefined) return null;
  return (
    <p className={parts.origin}>
      <span aria-hidden="true">
        <OriginMark origin={origin} variant="glyph" voice="silent" />
      </span>
      {originWords(origin)}
    </p>
  );
}

function Inspect(props: RunPanelProps) {
  const entity = props.state.selected;
  if (entity === null)
    return <p className={parts.muted}>Select an instance on the graph to see every channel at the playback time.</p>;
  return (
    <>
      <section className={parts.section} aria-label={`Observed · ${entity}`}>
        <h2 className={shared.sectionTitle}>
          {"Observed · "}
          <span className={parts.asIs}>{entity}</span>
          <span className={parts.note}>{`at ${runClock(props.state.positionNs, startOf(props))}`}</span>
        </h2>
        <OriginLine origin={props.state.origins.get(entity)} />
        <Observed review={props.review} state={props.state} plant={props.plant} entity={entity} />
      </section>
      {props.state.report.phase === "ready" && (
        <ReportPanel
          report={props.state.report.value}
          unit={props.unit}
          entity={entity}
          only="kpi"
          verify={props.state.verified}
        />
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
  const { ref, cut } = useCutBottom<HTMLElement>();
  return (
    <aside
      ref={ref}
      className={cx(styles.report, shared.scrolls, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-label="Run details"
      tabIndex={0}
    >
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
      <div className={styles.body} role="region" aria-label={tab === "report" ? "Report" : "Inspect"}>
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
  if (state.report.phase === "ready") {
    const { picked, positionNs } = state;
    const shown = picked !== null && picked.atNs === positionNs ? picked : null;
    return <ReportPanel report={state.report.value} unit={unit} entity={null} picked={shown} verify={state.verified} />;
  }
  if (state.report.phase === "loading")
    return (
      <p className={parts.muted} role="status">
        Reading the run…
      </p>
    );
  return (
    <p className={styles.failure} role="alert">
      {failureLine(state.report.failure.diagnostic)}
    </p>
  );
}
