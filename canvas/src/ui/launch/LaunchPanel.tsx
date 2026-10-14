import { Play, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { LaunchState, LaunchStore } from "../../app/launchStore";
import { clockOf } from "../../model/launch";
import type { Plant } from "../../model/plant";
import { useStore } from "../useStore";
import type { Draft } from "./draft";
import { draftOf, firstDraft } from "./draft";
import type { Change } from "./FaultRows";
import { LaunchForm } from "./LaunchForm";
import { useFocusOnMount } from "../useFocusOnMount";
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

function Progress({
  state,
  launch,
  title,
}: {
  readonly state: Running;
  readonly launch: LaunchStore;
  readonly title: string;
}) {
  const cancel = useFocusOnMount<HTMLButtonElement>();
  const words = state.worldId === null ? "Creating the world…" : `Running ${title} as ${state.worldId}`;
  return (
    <div className={cx(panel.body, styles.progress)}>
      <p className={styles.status} role="status">
        {state.canceling ? "Canceling at the next step…" : words}
      </p>
      <progress max={state.durationS} value={state.atS} aria-label="Simulated time" />
      <p className={styles.clock}>{`${clockOf(state.atS)} of ${clockOf(state.durationS)} simulated`}</p>
      <button
        type="button"
        ref={cancel}
        className={panel.secondary}
        aria-disabled={state.canceling}
        onClick={() => {
          launch.cancel();
        }}
      >
        Cancel the run
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
  const back = useFocusOnMount<HTMLButtonElement>();
  return (
    <div className={cx(panel.body, styles.failure)} role="alert">
      <p>
        <strong>The run did not finish.</strong> {state.failure.message}
      </p>
      <button
        type="button"
        ref={back}
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

interface BodyProps {
  readonly state: LaunchState;
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly draft: Draft;
  readonly change: (change: Change) => void;
}

function Body({ state, launch, plant, draft, change }: BodyProps) {
  if (state.phase === "running") return <Progress state={state} launch={launch} title={plant.title || plant.id} />;
  if (state.phase === "failed") return <Failure state={state} launch={launch} />;
  return (
    <LaunchForm
      plant={plant}
      draft={draft}
      change={change}
      onStart={(plan) => {
        void launch.start(plant.id, plan);
      }}
    />
  );
}

function useLaunch(launch: LaunchStore, onFinished: (worldId: string) => void) {
  const state = useStore(launch);
  useEffect(() => {
    if (launch.snapshot().phase !== "running") launch.reset();
    return () => {
      launch.cancel();
    };
  }, [launch]);
  useEffect(() => {
    if (state.phase === "done") {
      launch.reset();
      onFinished(state.worldId);
    }
  }, [state, launch, onFinished]);
  return state;
}

export function LaunchPanel({ launch, plant, onClose, onFinished }: LaunchPanelProps) {
  const state = useLaunch(launch, onFinished);
  const [draft, setDraft] = useState<Draft>(() => {
    const proposed = launch.proposal(plant.id);
    return proposed === null ? firstDraft : draftOf(proposed);
  });
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
      <Body state={state} launch={launch} plant={plant} draft={draft} change={setDraft} />
    </aside>
  );
}
