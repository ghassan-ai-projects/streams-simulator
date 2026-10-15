import { Dices, TriangleAlert } from "lucide-react";
import type { KeyboardEvent, SyntheticEvent } from "react";
import { useEffect, useRef, useState } from "react";
import type { LaunchPlan, PlanProblem } from "../../model/launch";
import { disturbanceWords, planProblems, watchWords } from "../../model/launch";
import { Pace } from "./PaceChoice";
import type { Plant } from "../../model/plant";
import { effectorOf } from "../../model/plant";
import { durationOf } from "../../model/workspace";
import type { Draft } from "./draft";
import { planOf } from "./draft";
import { CallRows } from "./CallRows";
import type { Change } from "./FaultRows";
import { FaultRows } from "./FaultRows";
import { cx } from "../cx";
import shared from "../shared.module.css";
import { useCutBottom } from "../useCutBottom";
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
  const text = problems.find((problem) => problem.field === field)?.text ?? "";
  return (
    <small id={id} className={cx(styles.problem, styles.inField)}>
      {text !== "" && <TriangleAlert size={13} aria-hidden="true" />}
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
            : "The true results stay hidden from the outside system while it runs, so it can be scored; they are shown when the run ends."}
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

function problemWords(problems: readonly PlanProblem[]): string {
  if (problems.length === 0) return "";
  return `${String(problems.length)} to fix before starting: ${problems.map((problem) => problem.text).join(" ")}`;
}

function startWords(plan: LaunchPlan): string {
  const length = durationOf(plan.durationS) || "the chosen time";
  const pace = plan.pace === "watch" ? `paced to take ${watchWords(plan.durationS)}` : "as fast as it can";
  return `Start simulates ${length} of plant time in ${plan.explore ? "Explore" : "Evaluate"} mode with ${disturbanceWords(plan)}, ${pace}, then saves the run and opens its review.`;
}

function fieldName(problem: PlanProblem): string {
  if (problem.field === "seed") return "Seed";
  if (problem.field === "duration") return "Duration";
  if (problem.field === "call") return `Effector call ${String((problem.call ?? 0) + 1)}`;
  return `Fault ${String(problem.field + 1)}`;
}

function planCue(shown: readonly PlanProblem[], plan: LaunchPlan): string {
  if (shown.length === 0) return startWords(plan);
  const names = shown.map(fieldName).join(", ");
  return shown.length === 1 ? `${names} needs attention before starting.` : `${names} need attention before starting.`;
}

function holdEnter(event: KeyboardEvent<HTMLFormElement>) {
  if (event.key === "Enter" && event.target instanceof HTMLSelectElement) event.preventDefault();
}

function PlanFields({
  plant,
  draft,
  selected,
  change,
  problems,
  shown,
  plan,
}: Pick<FormProps, "plant" | "draft" | "selected" | "change"> & {
  readonly problems: readonly PlanProblem[];
  readonly shown: readonly PlanProblem[];
  readonly plan: LaunchPlan;
}) {
  const { ref, cut } = useCutBottom<HTMLDivElement>([plant, draft, problems]);
  return (
    <div ref={ref} className={cx(styles.scroll, shared.cut)} data-scroll="run-plan" data-cut={cut ? "true" : undefined}>
      <fieldset className={panel.group}>
        <legend>Run</legend>
        <Seed draft={draft} change={change} problems={shown} />
        <Duration draft={draft} change={change} problems={shown} />
      </fieldset>
      <Mode draft={draft} change={change} />
      <Pace draft={draft} change={change} durationS={plan.durationS} />
      <FaultRows plant={plant} draft={draft} problems={shown} selected={selected} change={change} />
      <CallRows plant={plant} draft={draft} problems={shown} selected={selected} change={change} />
    </div>
  );
}

export function LaunchForm({ plant, draft, selected, change, onStart }: FormProps) {
  const [tried, setTried] = useState(0);
  const [told, setTold] = useState("");
  const form = useFocusFirstProblem(tried);
  const plan = planOf(draft, plant);
  const problems = planProblems(plan, (entity, effector) => effectorOf(plant, entity, effector));
  const shown = tried > 0 ? problems : [];
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    setTried((count) => count + 1);
    setTold(problemWords(problems));
    if (problems.length === 0) onStart(plan);
  };
  return (
    <form ref={form} className={styles.plan} onSubmit={submit} onKeyDown={holdEnter} noValidate aria-label="Run plan">
      <PlanFields
        plant={plant}
        draft={draft}
        selected={selected}
        change={change}
        problems={problems}
        shown={shown}
        plan={plan}
      />
      <div className={styles.footer}>
        <p id="launch-summary" className={styles.summary}>
          {planCue(shown, plan)}
        </p>
        <p className="sr-only" role="status">
          {told}
        </p>
        <button type="submit" className={panel.primary} aria-describedby="launch-summary">
          Start the run
        </button>
      </div>
    </form>
  );
}
