import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { LaunchPlan } from "../../model/launch";
import { planProblems } from "../../model/launch";
import type { Plant } from "../../model/plant";
import type { Draft } from "./draft";
import { firstDraft, planOf } from "./draft";
import type { Change } from "./FaultRows";
import { FaultRows } from "./FaultRows";
import styles from "./LaunchPanel.module.css";

interface FormProps {
  readonly plant: Plant;
  readonly onStart: (plan: LaunchPlan) => void;
}

interface DraftProps {
  readonly draft: Draft;
  readonly change: (change: Change) => void;
}

function Duration({ draft, change }: DraftProps) {
  return (
    <div className={styles.field}>
      <label htmlFor="launch-duration">Duration</label>
      <span className={styles.pair}>
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

function Settings({ draft, change }: DraftProps) {
  return (
    <fieldset className={styles.group}>
      <legend>Run</legend>
      <label className={styles.field}>
        <span>Seed</span>
        <input
          inputMode="numeric"
          value={draft.seed}
          onChange={(event) => {
            change((d) => ({ ...d, seed: event.target.value }));
          }}
        />
      </label>
      <Duration draft={draft} change={change} />
      <label className={styles.check}>
        <input
          type="checkbox"
          checked={draft.explore}
          onChange={(event) => {
            change((d) => ({ ...d, explore: event.target.checked }));
          }}
        />
        <span>Explore: truth visible after the run (unblinded, not for scoring)</span>
      </label>
    </fieldset>
  );
}

function Problems({ problems }: { readonly problems: readonly string[] }) {
  if (problems.length === 0) return null;
  return (
    <ul className={styles.problems} id="launch-problems">
      {problems.map((problem) => (
        <li key={problem}>{problem}</li>
      ))}
    </ul>
  );
}

export function LaunchForm({ plant, onStart }: FormProps) {
  const [draft, setDraft] = useState<Draft>(firstDraft);
  const plan = planOf(draft);
  const problems = planProblems(plan);
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (problems.length === 0) onStart(plan);
  };
  return (
    <form className={styles.form} onSubmit={submit} noValidate>
      <Settings draft={draft} change={setDraft} />
      <FaultRows plant={plant} draft={draft} change={setDraft} />
      <Problems problems={problems} />
      <button
        type="submit"
        className={styles.primary}
        aria-disabled={problems.length > 0}
        aria-describedby={problems.length > 0 ? "launch-problems" : undefined}
      >
        Start the run
      </button>
    </form>
  );
}
