import { Play, X } from "lucide-react";
import { useEffect } from "react";
import type { LaunchState, LaunchStore } from "../../app/launchStore";
import { clockOf } from "../../model/launch";
import type { Plant } from "../../model/plant";
import { useStore } from "../useStore";
import { LaunchForm } from "./LaunchForm";
import { cx } from "../cx";
import panel from "../panel.module.css";
import styles from "./LaunchPanel.module.css";

interface LaunchPanelProps {
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly onClose: () => void;
  readonly onFinished: (worldId: string) => void;
}

type Running = Extract<LaunchState, { phase: "running" }>;

function Progress({ state, launch }: { readonly state: Running; readonly launch: LaunchStore }) {
  return (
    <div className={cx(panel.body, styles.progress)}>
      <p className={styles.status} role="status">
        {state.worldId === null ? "Creating the world…" : `Running ${state.worldId}`}
      </p>
      <progress max={state.durationS} value={state.atS} aria-label="Simulated time" />
      <p className={styles.clock}>{`${clockOf(state.atS)} of ${clockOf(state.durationS)} simulated`}</p>
      <button
        type="button"
        className={panel.secondary}
        onClick={() => {
          launch.cancel();
        }}
      >
        Cancel
      </button>
    </div>
  );
}

function Failure({
  state,
  launch,
}: {
  readonly state: Extract<LaunchState, { phase: "failed" }>;
  readonly launch: LaunchStore;
}) {
  return (
    <div className={cx(panel.body, styles.failure)} role="alert">
      <p>{`The run did not finish: ${state.failure.message}`}</p>
      <button
        type="button"
        className={panel.secondary}
        onClick={() => {
          launch.reset();
        }}
      >
        Change the plan
      </button>
    </div>
  );
}

function Body({
  state,
  launch,
  plant,
}: {
  readonly state: LaunchState;
  readonly launch: LaunchStore;
  readonly plant: Plant;
}) {
  if (state.phase === "running") return <Progress state={state} launch={launch} />;
  if (state.phase === "failed") return <Failure state={state} launch={launch} />;
  return (
    <LaunchForm
      plant={plant}
      onStart={(plan) => {
        void launch.start(plant.id, plan);
      }}
    />
  );
}

export function LaunchPanel({ launch, plant, onClose, onFinished }: LaunchPanelProps) {
  const state = useStore(launch);
  useEffect(() => {
    if (state.phase === "done") {
      launch.reset();
      onFinished(state.worldId);
    }
  }, [state, launch, onFinished]);
  return (
    <aside className={panel.panel} aria-labelledby="launch-title">
      <header className={panel.head}>
        <Play size={16} aria-hidden="true" />
        <h2 id="launch-title">{`Run ${plant.title || plant.id}`}</h2>
        <button
          type="button"
          className={panel.icon}
          aria-label="Close the run panel"
          disabled={state.phase === "running"}
          onClick={onClose}
        >
          <X size={16} aria-hidden="true" />
        </button>
      </header>
      <Body state={state} launch={launch} plant={plant} />
    </aside>
  );
}
