import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { LaunchPlan } from "../../model/launch";
import { planProblems } from "../../model/launch";
import type { Plant } from "../../model/plant";
import type { Draft } from "./draft";
import { planOf } from "./draft";
import type { Change } from "./FaultRows";
import { FaultRows } from "./FaultRows";
import { cx } from "../cx";
import panel from "../panel.module.css";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./LaunchPanel.module.css";

interface FormProps {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly change: (change: Change) => void;
  readonly onStart: (plan: LaunchPlan) => void;
}

interface DraftProps {
  readonly draft: Draft;
  readonly change: (change: Change) => void;
}

function Duration({ draft, change }: DraftProps) {
  return (
    <div className={panel.field}>
      <label htmlFor="launch-duration">Duration</label>
      <span className={cx(panel.pair, styles.pair)}>
        <input
          id="launch-duration"
          inputMode="numeric"
          value={draft.duration}
          onChange={(event) => {
            change((d) => ({ ...d, duration: event.target.value }));
          }}
        />
        <select
          aria-label="Duration unit"
          value={draft.unit}
          onChange={(event) => {
            change((d) => ({ ...d, unit: event.target.value === "hours" ? "hours" : "minutes" }));
          }}
        >
          <option value="minutes">minutes</option>
          <option value="hours">hours</option>
        </select>
      </span>
    </div>
  );
}

function Mode({ draft, change }: DraftProps) {
  const choose = (explore: boolean) => {
    change((d) => ({ ...d, explore }));
  };
  return (
    <fieldset className={panel.group}>
      <legend>Mode</legend>
      <label className={panel.check}>
        <input
          type="radio"
          name="launch-mode"
          checked={draft.explore}
          onChange={() => {
            choose(true);
          }}
        />
        <span>Explore: the truth is visible after the run; not for scoring.</span>
      </label>
      <label className={panel.check}>
        <input
          type="radio"
          name="launch-mode"
          checked={!draft.explore}
          onChange={() => {
            choose(false);
          }}
        />
        <span>
          Evaluate: the truth stays sealed while it runs, so a consumer can be scored; it is revealed when the run ends.
        </span>
      </label>
    </fieldset>
  );
}

function Settings({ draft, change }: DraftProps) {
  const seed = useFocusOnMount<HTMLInputElement>();
  return (
    <fieldset className={panel.group}>
      <legend>Run</legend>
      <div className={panel.field}>
        <label htmlFor="launch-seed">Seed</label>
        <input
          id="launch-seed"
          ref={seed}
          inputMode="numeric"
          aria-describedby="launch-seed-hint"
          value={draft.seed}
          onChange={(event) => {
            change((d) => ({ ...d, seed: event.target.value }));
          }}
        />
        <small id="launch-seed-hint" className={styles.note}>
          The same seed gives the same run.
        </small>
      </div>
      <Duration draft={draft} change={change} />
    </fieldset>
  );
}

function Problems({ problems }: { readonly problems: readonly string[] }) {
  const list = useFocusOnMount<HTMLUListElement>();
  return (
    <ul className={panel.problems} id="launch-problems" ref={list} tabIndex={-1} aria-label="Fix before starting">
      {problems.map((problem) => (
        <li key={problem}>{problem}</li>
      ))}
    </ul>
  );
}

export function LaunchForm({ plant, draft, change, onStart }: FormProps) {
  const [tried, setTried] = useState(0);
  const plan = planOf(draft);
  const problems = planProblems(plan);
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    setTried((count) => count + 1);
    if (problems.length === 0) onStart(plan);
  };
  return (
    <form className={panel.body} onSubmit={submit} noValidate>
      <Settings draft={draft} change={change} />
      <Mode draft={draft} change={change} />
      <FaultRows plant={plant} draft={draft} change={change} />
      {tried > 0 && problems.length > 0 && <Problems key={tried} problems={problems} />}
      <button
        type="submit"
        className={panel.primary}
        aria-describedby={tried > 0 && problems.length > 0 ? "launch-problems" : undefined}
      >
        Start the run
      </button>
    </form>
  );
}
