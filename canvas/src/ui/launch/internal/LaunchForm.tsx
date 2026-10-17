import { holdEnter, panel, shared } from "../../primitives";
import type { SyntheticEvent } from "react";
import { useEffect, useRef, useState } from "react";
import type { Draft, LaunchPlan, PlanProblem } from "../../../model/launch";
import { planCue, planOf, planProblems, problemWords } from "../../../model/launch";
import { Pace } from "./PaceChoice";
import type { Plant } from "../../../model/plant";
import { effectorOf } from "../../../model/plant";
import { CallRows } from "./CallRows";
import type { Change } from "./FaultRows";
import { FaultRows } from "./FaultRows";
import { cx, useCutBottom } from "../../primitives";
import { Duration, Mode, Seed } from "./RunSettings";
import styles from "./LaunchPanel.module.css";

interface FormProps {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly selected: string | null;
  readonly change: (change: Change) => void;
  readonly onStart: (plan: LaunchPlan) => void;
}

function useFocusFirstProblem(tried: number) {
  const form = useRef<HTMLFormElement>(null);
  useEffect(() => {
    if (tried > 0) form.current?.querySelector<HTMLElement>("[aria-invalid='true']")?.focus();
  }, [tried]);
  return form;
}

function PlanFields({
  plant,
  draft,
  selected,
  change,
  shown,
  plan,
}: Pick<FormProps, "plant" | "draft" | "selected" | "change"> & {
  readonly shown: readonly PlanProblem[];
  readonly plan: LaunchPlan;
}) {
  const { ref, cut } = useCutBottom<HTMLDivElement>();
  return (
    <div
      ref={ref}
      className={cx(styles.scroll, shared.scrolls, shared.cut)}
      data-scroll="run-plan"
      data-cut={cut ? "true" : undefined}
    >
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
      <PlanFields plant={plant} draft={draft} selected={selected} change={change} shown={shown} plan={plan} />
      <div className={styles.footer}>
        <p id="launch-summary" className={styles.summary}>
          {planCue(shown, plan)}
        </p>
        <p className="sr-only" role="status">
          {told}
        </p>
        <button type="submit" className={cx(shared.primary, panel.primary)} aria-describedby="launch-summary">
          Start the run
        </button>
      </div>
    </form>
  );
}
