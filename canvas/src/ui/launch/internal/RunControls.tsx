import { Pause, Play, StepForward } from "lucide-react";
import type { LaunchState, LaunchStore } from "../../../app/launch";
import { pendingWords, speedChoices, speedValue } from "../../../model/launch";
import { cx, panel, shared } from "../../primitives";
import styles from "./RunControls.module.css";
import { ControlRefusal } from "./ControlRefusal";

export type Running = Extract<LaunchState, { phase: "running" }>;

export function controllable(state: Running): boolean {
  return state.worldId !== null && !state.stopping && state.control.pending === null;
}

function PauseToggle({ state, launch }: { readonly state: Running; readonly launch: LaunchStore }) {
  const Icon = state.paused ? Play : Pause;
  return (
    <button
      type="button"
      className={cx(shared.secondary, panel.secondary)}
      aria-disabled={!controllable(state)}
      onClick={() => {
        void (state.paused ? launch.continueRun() : launch.pause());
      }}
    >
      <Icon size={14} aria-hidden="true" />
      {state.paused ? "Resume" : "Pause"}
    </button>
  );
}

function Speed({ state, launch }: { readonly state: Running; readonly launch: LaunchStore }) {
  const live = controllable(state);
  return (
    <label className={styles.speed}>
      Speed
      <select
        value={speedValue(state.rate)}
        aria-disabled={!live}
        onChange={(event) => {
          if (live) void launch.setRate(Number(event.target.value));
        }}
      >
        {speedChoices(state.rate).map((choice) => (
          <option key={choice.value} value={choice.value}>
            {choice.label}
          </option>
        ))}
      </select>
    </label>
  );
}

export function RunControls({ state, launch }: { readonly state: Running; readonly launch: LaunchStore }) {
  return (
    <fieldset className={panel.group}>
      <legend>Controls</legend>
      <div className={styles.controls}>
        <PauseToggle state={state} launch={launch} />
        <button
          type="button"
          className={cx(shared.secondary, panel.secondary)}
          aria-disabled={!controllable(state) || !state.paused}
          onClick={() => {
            void launch.step();
          }}
        >
          <StepForward size={14} aria-hidden="true" />
          Step one chunk
        </button>
        <Speed state={state} launch={launch} />
      </div>
      <p className={styles.note} aria-live="polite">
        {pendingWords(state.control.pending)}
      </p>
      <ControlRefusal words={state.control.refusal} />
      <p className={panel.hint}>
        Pause stops the simulated clock only. A chunk is the stretch of simulated time the simulator advances in one go.
      </p>
    </fieldset>
  );
}
