import { Dices, TriangleAlert } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useEffect, useRef, useState } from "react";
import type { LaunchPlan, PlanProblem } from "../../model/launch";
import { planProblems } from "../../model/launch";
import type { Plant } from "../../model/plant";
import { durationOf } from "../../model/workspace";
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
  readonly selected: string | null;
  readonly change: (change: Change) => void;
  readonly onStart: (plan: LaunchPlan) => void;
}

interface DraftProps {
  readonly draft: Draft;
  readonly change: (change: Change) => void;
  readonly problems: readonly PlanProblem[];
}

function FieldProblem({
  id,
  problems,
  field,
}: { readonly id: string; readonly field: string } & Pick<DraftProps, "problems">) {
  const text = problems.find((problem) => problem.field === field)?.text;
  if (text === undefined) return null;
  return (
    <small id={id} className={styles.problem}>
      <TriangleAlert size={13} aria-hidden="true" />
      {text}
    </small>
  );
}

function described(problems: readonly PlanProblem[], field: string, id: string, hint = "") {
  const wrong = problems.some((problem) => problem.field === field);
  const ids = [hint, wrong ? id : ""].filter((part) => part !== "").join(" ");
  return { "aria-invalid": wrong || undefined, "aria-describedby": ids === "" ? undefined : ids };
}

function Duration({ draft, change, problems }: DraftProps) {
  return (
    <div className={panel.field}>
      <label htmlFor="launch-duration">Duration</label>
      <span className={cx(panel.pair, styles.pair)}>
        <input
          id="launch-duration"
          inputMode="decimal"
          value={draft.duration}
          {...described(problems, "duration", "launch-duration-problem")}
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
      <FieldProblem id="launch-duration-problem" field="duration" problems={problems} />
    </div>
  );
}

function ModeChoice({
  explore,
  checked,
  onChoose,
}: {
  readonly explore: boolean;
  readonly checked: boolean;
  readonly onChoose: () => void;
}) {
  return (
    <label className={styles.mode}>
      <input type="radio" name="launch-mode" checked={checked} onChange={onChoose} />
      <span>
        <strong>{explore ? "Explore" : "Evaluate"}</strong>
        <small>
          {explore
            ? "The truth is visible after the run; not for scoring."
            : "The truth stays sealed while it runs, so a consumer can be scored; it is revealed when the run ends."}
        </small>
      </span>
    </label>
  );
}

function Mode({ draft, change }: Omit<DraftProps, "problems">) {
  return (
    <fieldset className={panel.group}>
      <legend>Mode</legend>
      {[true, false].map((explore) => (
        <ModeChoice
          key={String(explore)}
          explore={explore}
          checked={draft.explore === explore}
          onChoose={() => {
            change((d) => ({ ...d, explore }));
          }}
        />
      ))}
    </fieldset>
  );
}

function Seed({ draft, change, problems }: DraftProps) {
  const seed = useFocusOnMount<HTMLInputElement>();
  return (
    <div className={panel.field}>
      <label htmlFor="launch-seed">Seed</label>
      <span className={styles.seed}>
        <input
          id="launch-seed"
          ref={seed}
          inputMode="numeric"
          value={draft.seed}
          {...described(problems, "seed", "launch-seed-problem", "launch-seed-hint")}
          onChange={(event) => {
            change((d) => ({ ...d, seed: event.target.value }));
          }}
        />
        <button
          type="button"
          className={panel.icon}
          aria-label="Pick a new seed"
          title="Pick a new seed"
          onClick={() => {
            change((d) => ({ ...d, seed: String(Math.floor(Math.random() * 1_000_000)) }));
          }}
        >
          <Dices size={16} aria-hidden="true" />
        </button>
      </span>
      <FieldProblem id="launch-seed-problem" field="seed" problems={problems} />
      <small id="launch-seed-hint" className={styles.note}>
        The same seed, plant, faults and simulator version give the same run.
      </small>
    </div>
  );
}

function useFocusFirstProblem(tried: number) {
  const form = useRef<HTMLFormElement>(null);
  useEffect(() => {
    if (tried > 0) form.current?.querySelector<HTMLElement>("[aria-invalid='true']")?.focus();
  }, [tried]);
  return form;
}

function startWords(plan: LaunchPlan): string {
  const length = durationOf(plan.durationS) || "the chosen time";
  const faults = plan.faults.length;
  const injected = faults === 0 ? "no faults" : `${String(faults)} ${faults === 1 ? "fault" : "faults"}`;
  return `Start simulates ${length} of plant time in ${plan.explore ? "Explore" : "Evaluate"} mode with ${injected}, as fast as it can, then saves the run and opens its review.`;
}

export function LaunchForm({ plant, draft, selected, change, onStart }: FormProps) {
  const [tried, setTried] = useState(0);
  const form = useFocusFirstProblem(tried);
  const plan = planOf(draft);
  const problems = planProblems(plan);
  const shown = tried > 0 ? problems : [];
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    setTried((count) => count + 1);
    if (problems.length === 0) onStart(plan);
  };
  return (
    <form ref={form} className={panel.body} onSubmit={submit} noValidate aria-label="Run plan">
      <fieldset className={panel.group}>
        <legend>Run</legend>
        <Seed draft={draft} change={change} problems={shown} />
        <Duration draft={draft} change={change} problems={shown} />
      </fieldset>
      <Mode draft={draft} change={change} />
      <FaultRows plant={plant} draft={draft} problems={shown} selected={selected} change={change} />
      <p id="launch-summary" className={styles.summary}>
        {startWords(plan)}
      </p>
      <p className="sr-only" role="status">
        {shown.length === 0
          ? ""
          : `${String(shown.length)} to fix before starting: ${shown.map((p) => p.text).join(" ")}`}
      </p>
      <button type="submit" className={panel.primary} aria-describedby="launch-summary">
        Start the run
      </button>
    </form>
  );
}
