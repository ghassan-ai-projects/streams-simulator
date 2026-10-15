import { Play, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { LaunchState, LaunchStore, Proposal } from "../../app/launchStore";
import type { CallOutcome } from "../../model/launch";
import { clockOf, keepKnownFaults, planSummary, speedOf } from "../../model/launch";
import { humanise } from "../../model/report";
import { durationOf } from "../../model/workspace";
import type { Plant } from "../../model/plant";
import { faultsOf } from "../../model/plant";
import { useStore } from "../useStore";
import type { Draft } from "./draft";
import { draftOf, firstDraft, pickedOnGraph } from "./draft";
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
  readonly selected?: string | null;
  readonly onClose: () => void;
  readonly onFinished: (worldId: string) => void;
}

type Running = Extract<LaunchState, { phase: "running" }>;
type Done = Extract<LaunchState, { phase: "done" }>;

function outcomeWords(outcome: CallOutcome): string {
  const call = `${clockOf(outcome.atS)} · ${outcome.entity} · ${humanise(outcome.effector)}`;
  return outcome.accepted ? `${call}: accepted` : `${call}: command not accepted: ${outcome.reason}`;
}

function Result({ state, onOpen }: { readonly state: Done; readonly onOpen: () => void }) {
  const status = useFocusOnMount<HTMLParagraphElement>();
  const refused = state.calls.filter((call) => !call.accepted).length;
  return (
    <div className={panel.body}>
      <p className={styles.status} role="status" ref={status} tabIndex={-1}>
        {`Run ${state.worldId} ${state.early ? "stopped early and was saved" : "finished"}. ${
          refused === 0
            ? "Every effector call was accepted."
            : `${String(refused)} of ${String(state.calls.length)} effector calls were not accepted.`
        }`}
      </p>
      <ul className={styles.outcomes} aria-label="Effector call results">
        {state.calls.map((outcome, index) => (
          <li key={String(index)}>{outcomeWords(outcome)}</li>
        ))}
      </ul>
      <button type="button" className={panel.primary} onClick={onOpen}>
        Open the review
      </button>
    </div>
  );
}

function ProgressBar({ state }: { readonly state: Running }) {
  return (
    <div className={styles.bar}>
      <progress max={state.durationS} value={state.atS} aria-label="Simulated time" />
      {[
        ...state.plan.faults.map((fault) => ({ ...fault, what: fault.fault })),
        ...state.plan.calls.map((call) => ({ ...call, what: call.effector })),
      ].map((fault, index) => (
        <span
          key={String(index)}
          className={fault.atS <= state.atS ? styles.tickDone : styles.tick}
          style={{ left: `${String((fault.atS / state.durationS) * 100)}%` }}
          title={`${fault.entity} ${humanise(fault.what)} at ${clockOf(fault.atS)}`}
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
  const speed = speedOf(state.atS, state.atMs - state.startedMs);
  return (
    <div className={cx(panel.body, styles.progress)}>
      <p className={styles.status} role="status" ref={status} tabIndex={-1}>
        {state.stopping ? "Stopping at the next step…" : words}
      </p>
      <p className={styles.clock}>{planSummary(state.plan, durationOf(state.durationS))}</p>
      <ProgressBar state={state} />
      <p className={styles.clock}>
        {`${clockOf(state.atS)} of ${clockOf(state.durationS)} simulated${speed === "" ? "" : ` · ${speed}`}`}
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
  readonly selected: string | null;
  readonly change: (change: Change) => void;
  readonly onFinished: (worldId: string) => void;
}

function Elsewhere({ state }: { readonly state: Running }) {
  return (
    <div className={panel.body}>
      <p className={styles.status} role="status">{`A run of ${state.origin.title} is in progress.`}</p>
      <p className={panel.hint}>One run at a time: follow it from the top bar, or cancel it on its plant.</p>
    </div>
  );
}

function Body({ state, launch, plant, path, draft, selected, change, onFinished }: BodyProps) {
  const own = state.phase !== "idle" && state.origin.domain === plant.id;
  if (state.phase === "running") {
    return own ? (
      <Progress state={state} launch={launch} title={plant.title || plant.id} />
    ) : (
      <Elsewhere state={state} />
    );
  }
  if (state.phase === "failed" && own) return <Failure state={state} launch={launch} />;
  if (state.phase === "done" && own && state.calls.length > 0)
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

function useLaunch(launch: LaunchStore, plant: string, onFinished: (worldId: string) => void) {
  const state = useStore(launch);
  const [watched, setWatched] = useState(false);
  if (!watched && state.phase === "running" && state.origin.domain === plant) setWatched(true);
  useEffect(() => {
    if (watched && state.phase === "done" && state.origin.domain === plant && state.calls.length === 0) {
      launch.reset();
      onFinished(state.worldId);
    }
  }, [watched, state, launch, plant, onFinished]);
  return state;
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
  const state = useLaunch(launch, plant.id, onFinished);
  const { proposed, draft, setDraft } = useDraft(launch, plant, selected);
  return (
    <aside className={panel.panel} aria-labelledby="launch-title">
      <header className={panel.head}>
        <Play size={16} aria-hidden="true" />
        <h2 id="launch-title">{`Run ${plant.title || plant.id}`}</h2>
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
      />
    </aside>
  );
}
