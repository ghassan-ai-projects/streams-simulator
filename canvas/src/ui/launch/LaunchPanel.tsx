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
  readonly path: string;
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
  const status = useFocusOnMount<HTMLParagraphElement>();
  const words = state.worldId === null ? "Creating the world…" : `Running ${title} as ${state.worldId}`;
  return (
    <div className={cx(panel.body, styles.progress)}>
      <p className={styles.status} role="status" ref={status} tabIndex={-1}>
        {state.stopping ? "Stopping at the next step…" : words}
      </p>
      <progress max={state.durationS} value={state.atS} aria-label="Simulated time" />
      <p className={styles.clock}>{`${clockOf(state.atS)} of ${clockOf(state.durationS)} simulated`}</p>
      <button
        type="button"
        className={panel.secondary}
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
        {state.kept !== null && ` What ran before the failure is kept as run ${state.kept}.`}
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
  readonly path: string;
  readonly draft: Draft;
  readonly change: (change: Change) => void;
}

function Elsewhere({ state }: { readonly state: Running }) {
  return (
    <div className={panel.body}>
      <p className={styles.status} role="status">{`A run of ${state.origin.title} is in progress.`}</p>
      <p className={panel.hint}>One run at a time: follow it from the top bar, or cancel it on its plant.</p>
    </div>
  );
}

function Body({ state, launch, plant, path, draft, change }: BodyProps) {
  const own = state.phase !== "idle" && state.origin.domain === plant.id;
  if (state.phase === "running") {
    return own ? (
      <Progress state={state} launch={launch} title={plant.title || plant.id} />
    ) : (
      <Elsewhere state={state} />
    );
  }
  if (state.phase === "failed" && own) return <Failure state={state} launch={launch} />;
  return (
    <LaunchForm
      plant={plant}
      draft={draft}
      change={change}
      onStart={(plan) => {
        void launch.start({ domain: plant.id, title: plant.title || plant.id, path }, plan);
      }}
    />
  );
}

function useLaunch(launch: LaunchStore, plant: string, onFinished: (worldId: string) => void) {
  const state = useStore(launch);
  const [watched, setWatched] = useState(false);
  if (!watched && state.phase === "running" && state.origin.domain === plant) setWatched(true);
  useEffect(() => {
    if (watched && state.phase === "done" && state.origin.domain === plant) {
      launch.reset();
      onFinished(state.worldId);
    }
  }, [watched, state, launch, plant, onFinished]);
  return state;
}

export function LaunchPanel({ launch, plant, path, onClose, onFinished }: LaunchPanelProps) {
  const state = useLaunch(launch, plant.id, onFinished);
  const [proposed] = useState(() => launch.proposal(plant.id));
  const [draft, setDraft] = useState<Draft>(() => (proposed === null ? firstDraft : draftOf(proposed.plan)));
  return (
    <aside className={panel.panel} aria-labelledby="launch-title">
      <header className={panel.head}>
        <Play size={16} aria-hidden="true" />
        <h2 id="launch-title">{`Run ${plant.title || plant.id}`}</h2>
        <button type="button" className={panel.icon} aria-label="Close the run panel" onClick={onClose}>
          <X size={16} aria-hidden="true" />
        </button>
      </header>
      {proposed !== null && proposed.left.length > 0 && state.phase !== "running" && (
        <p className={panel.hint} role="note">
          {`Copied from the earlier run: its seed, length, mode and faults. Not copied: ${proposed.left.join("; ")}.`}
        </p>
      )}
      <Body state={state} launch={launch} plant={plant} path={path} draft={draft} change={setDraft} />
    </aside>
  );
}
