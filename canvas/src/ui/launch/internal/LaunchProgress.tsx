import { panel, shared, cx, useFocusOnMount } from "../../primitives";
import type { LaunchState, LaunchStore } from "../../../app/launch";
import { planSummary, progressTicks, runProgressWords, runStageWords, runTitle } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { plantTitle } from "../../../model/plant";
import { durationOf } from "../../../model/shared/duration";
import { ActNow } from "./ActNow";
import { RunControls } from "./RunControls";
import styles from "./LaunchPanel.module.css";

export type Running = Extract<LaunchState, { phase: "running" }>;

function ProgressBar({ state }: { readonly state: Running }) {
  return (
    <div className={styles.bar}>
      <progress max={state.durationS} value={state.atS} aria-label="Simulated time" />
      {progressTicks(state.plan).map((mark, index) => (
        <span
          key={String(index)}
          className={cx(mark.atS <= state.atS ? styles.tickDone : styles.tick, mark.call && styles.tickCall)}
          style={{ left: `${String((mark.atS / state.durationS) * 100)}%` }}
          title={mark.title}
          aria-hidden="true"
        />
      ))}
    </div>
  );
}

export function Progress({
  state,
  launch,
  plant,
}: {
  readonly state: Running;
  readonly launch: LaunchStore;
  readonly plant: Plant;
}) {
  const status = useFocusOnMount<HTMLParagraphElement>();
  const title = runTitle(state.plan, plantTitle(plant));
  const words = state.worldId === null ? "Creating the world…" : `Running ${title} as run ${state.worldId}`;
  return (
    <div className={cx(panel.body, styles.progress)}>
      <p className={styles.status} role="status" ref={status} tabIndex={-1}>
        {state.stopping ? "Stopping at the next step…" : words}
      </p>
      {state.plan.purpose !== undefined && <p className={styles.clock}>{state.plan.purpose}</p>}
      <p className={styles.clock}>{planSummary(state.plan, durationOf(state.durationS))}</p>
      <ProgressBar state={state} />
      <p className={styles.clock}>{`${runStageWords(state)} · ${runProgressWords(state, " simulated")}`}</p>
      <RunControls state={state} launch={launch} />
      <ActNow state={state} launch={launch} plant={plant} />
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        aria-disabled={state.stopping}
        onClick={() => {
          launch.stopEarly();
        }}
      >
        Stop now
      </button>
      <p className={panel.hint}>
        Stop now ends the run at the next step and keeps what ran as a shorter run you can review. You can close this
        panel or leave the plant: the run goes on and its progress shows in the top bar.
      </p>
    </div>
  );
}

export function Elsewhere({ state }: { readonly state: Running }) {
  return (
    <div className={panel.body}>
      <p className={styles.status} role="status">{`A run of ${state.origin.title} is in progress.`}</p>
      <p className={panel.hint}>One run at a time: follow it from the top bar, or cancel it on its plant.</p>
    </div>
  );
}
