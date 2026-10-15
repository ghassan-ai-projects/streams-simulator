import { Play, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { LaunchState, LaunchStore, Proposal } from "../../app/launchStore";
import { clockOf, keepKnownFaults, planSummary, rateWords } from "../../model/launch";
import { humanise } from "../../model/report";
import { durationOf } from "../../model/workspace";
import type { Plant } from "../../model/plant";
import { faultsOf } from "../../model/plant";
import { useStore } from "../useStore";
import type { Draft } from "./draft";
import { draftOf, firstDraft, pickedOnGraph } from "./draft";
import type { Change } from "./FaultRows";
import { LaunchForm } from "./LaunchForm";
import { Result } from "./LaunchResult";
import { RunningMark } from "./RunningMark";
import { useFocusOnMount } from "../useFocusOnMount";
import { cx } from "../cx";
import panel from "../panel.module.css";
import styles from "./LaunchPanel.module.css";

interface LaunchPanelProps {
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly path: string;
  readonly selected?: string | null;
  readonly onClose: () => void;
  readonly onFinished: (worldId: string) => void;
}

type Running = Extract<LaunchState, { phase: "running" }>;
type Done = Extract<LaunchState, { phase: "done" }>;

function ProgressBar({ state }: { readonly state: Running }) {
  const marks = [
    ...state.plan.faults.map((fault) => ({ ...fault, what: fault.fault, call: false })),
    ...state.plan.calls.map((call) => ({ ...call, what: call.effector, call: true })),
  ];
  return (
    <div className={styles.bar}>
      <progress max={state.durationS} value={state.atS} aria-label="Simulated time" />
      {marks.map((mark, index) => (
        <span
          key={String(index)}
          className={cx(mark.atS <= state.atS ? styles.tickDone : styles.tick, mark.call && styles.tickCall)}
          style={{ left: `${String((mark.atS / state.durationS) * 100)}%` }}
          title={`${mark.entity} ${humanise(mark.what)} at ${clockOf(mark.atS)}`}
          aria-hidden="true"
        />
      ))}
    </div>
  );
}

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
  const words = state.worldId === null ? "Creating the world…" : `Running ${title} as run ${state.worldId}`;
  const speed = rateWords(state.rate);
  return (
    <div className={cx(panel.body, styles.progress)}>
      <p className={styles.status} role="status" ref={status} tabIndex={-1}>
        {state.stopping ? "Stopping at the next step…" : words}
      </p>
      <p className={styles.clock}>{planSummary(state.plan, durationOf(state.durationS))}</p>
      <ProgressBar state={state} />
      <p className={styles.clock}>
        {`Running · ${clockOf(state.atS)} of ${clockOf(state.durationS)} simulated${speed === "" ? "" : ` · ${speed}`}`}
      </p>
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
        <strong>The run did not finish.</strong> {state.words}
        {state.kept !== null && ` What ran before the failure is kept as run ${state.kept}.`}
      </p>
      <details className={styles.detail}>
        <summary>Binary message</summary>
        <p>{state.failure.message}</p>
      </details>
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
  readonly selected: string | null;
  readonly change: (change: Change) => void;
  readonly onFinished: (worldId: string) => void;
  readonly watched: boolean;
}

function Elsewhere({ state }: { readonly state: Running }) {
  return (
    <div className={panel.body}>
      <p className={styles.status} role="status">{`A run of ${state.origin.title} is in progress.`}</p>
      <p className={panel.hint}>One run at a time: follow it from the top bar, or cancel it on its plant.</p>
    </div>
  );
}

function showsResult(state: LaunchState, own: boolean, watched: boolean): state is Done {
  return state.phase === "done" && own && (state.calls.length > 0 || watched);
}

function Body({ state, launch, plant, path, draft, selected, change, onFinished, watched }: BodyProps) {
  const own = state.phase !== "idle" && state.origin.domain === plant.id;
  if (state.phase === "running") {
    return own ? (
      <Progress state={state} launch={launch} title={plant.title || plant.id} />
    ) : (
      <Elsewhere state={state} />
    );
  }
  if (state.phase === "failed" && own) return <Failure state={state} launch={launch} />;
  if (showsResult(state, own, watched))
    return (
      <Result
        state={state}
        onOpen={() => {
          launch.reset();
          onFinished(state.worldId);
        }}
      />
    );
  return (
    <LaunchForm
      plant={plant}
      draft={draft}
      selected={selected}
      change={change}
      onStart={(plan) => {
        void launch.start({ domain: plant.id, title: plant.title || plant.id, path }, plan);
      }}
    />
  );
}

const finishedHold = 2000;

function useLaunch(launch: LaunchStore, plant: string, onFinished: (worldId: string) => void) {
  const state = useStore(launch);
  const [watched, setWatched] = useState(false);
  if (!watched && state.phase === "running" && state.origin.domain === plant) setWatched(true);
  useEffect(() => {
    if (!watched || state.phase !== "done" || state.origin.domain !== plant || state.calls.length > 0) return undefined;
    const timer = window.setTimeout(() => {
      launch.reset();
      onFinished(state.worldId);
    }, finishedHold);
    return () => {
      window.clearTimeout(timer);
    };
  }, [watched, state, launch, plant, onFinished]);
  return { state, watched };
}

function shownProposal(proposed: Proposal): boolean {
  return proposed.from !== "" || proposed.left.length > 0;
}

function proposalNote(proposed: Proposal): string {
  const source = proposed.from === "" ? "the earlier run" : proposed.from;
  const copied = `Prefilled from ${source}: seed, duration, mode and faults. Change them if needed.`;
  return proposed.left.length === 0 ? copied : `${copied} Not copied: ${proposed.left.join("; ")}.`;
}

function prefillOf(proposal: Proposal | null, plant: Plant): Proposal | null {
  if (proposal === null) return null;
  const kept = keepKnownFaults(proposal.plan, (entity) => faultsOf(plant, entity));
  return { ...proposal, plan: kept.plan, left: [...proposal.left, ...kept.dropped] };
}

function useDraft(launch: LaunchStore, plant: Plant, selected: string | null) {
  const [proposed] = useState(() => prefillOf(launch.proposal(plant.id), plant));
  useEffect(
    () => () => {
      launch.forgetProposal();
    },
    [launch],
  );
  const [draft, setDraft] = useState<Draft>(() => (proposed === null ? firstDraft : draftOf(proposed.plan)));
  const [seen, setSeen] = useState(selected);
  if (seen !== selected) {
    setSeen(selected);
    if (selected !== null) setDraft((current) => pickedOnGraph(current, selected));
  }
  return { proposed, draft, setDraft };
}

export function LaunchPanel({ launch, plant, path, selected = null, onClose, onFinished }: LaunchPanelProps) {
  const { state, watched } = useLaunch(launch, plant.id, onFinished);
  const { proposed, draft, setDraft } = useDraft(launch, plant, selected);
  return (
    <aside className={panel.panel} aria-labelledby="launch-title">
      <header className={panel.head}>
        <Play size={16} aria-hidden="true" />
        <h2 id="launch-title" className={styles.title}>{`Run ${plant.title || plant.id}`}</h2>
        {state.phase === "running" && state.origin.domain === plant.id && <RunningMark state={state} />}
        <button type="button" className={panel.icon} aria-label="Close the run panel" onClick={onClose}>
          <X size={16} aria-hidden="true" />
        </button>
      </header>
      {proposed !== null && shownProposal(proposed) && state.phase !== "running" && (
        <p className={panel.hint} role="note">
          {proposalNote(proposed)}
        </p>
      )}
      <Body
        state={state}
        launch={launch}
        plant={plant}
        path={path}
        draft={draft}
        selected={selected}
        change={setDraft}
        onFinished={onFinished}
        watched={watched}
      />
    </aside>
  );
}
