import { holdEnter, panel, shared } from "../../primitives";
import type { ReactNode, SyntheticEvent } from "react";
import { useEffect, useRef, useState } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { Draft, LaunchPlan, PlanProblem } from "../../../model/launch";
import { agendaOf, faultCategoryOf, planCue, planOf, planProblems, problemWords } from "../../../model/launch";
import { Pace } from "./PaceChoice";
import type { Plant } from "../../../model/plant";
import { effectorOf, plantTitle } from "../../../model/plant";
import { CallRows } from "./CallRows";
import type { Change } from "./FaultRows";
import { FaultRows } from "./FaultRows";
import { cx, useCutBottom } from "../../primitives";
import { DeliveryRows } from "./DeliveryRows";
import { PlanAgenda } from "./PlanAgenda";
import { PlanIdentity } from "./PlanIdentity";
import { Readiness } from "./Readiness";
import { Duration, Mode, Seed } from "./RunSettings";
import { usePlanCatalog } from "./usePlanCatalog";
import styles from "./LaunchPanel.module.css";

interface FormProps {
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly draft: Draft;
  readonly selected: string | null;
  readonly change: (change: Change) => void;
  readonly onStart: (plan: LaunchPlan) => void;
  readonly note: ReactNode;
}

function useFocusFirstProblem(tried: number) {
  const form = useRef<HTMLFormElement>(null);
  useEffect(() => {
    if (tried > 0) form.current?.querySelector<HTMLElement>("[aria-invalid='true']")?.focus();
  }, [tried]);
  return form;
}

type Catalog = ReturnType<typeof usePlanCatalog>;

interface FieldsProps extends Pick<FormProps, "plant" | "draft" | "selected" | "change" | "note"> {
  readonly shown: readonly PlanProblem[];
  readonly plan: LaunchPlan;
  readonly catalog: Catalog;
}

function PlanFields({ plant, draft, selected, change, shown, plan, note, catalog }: FieldsProps) {
  const { ref, cut } = useCutBottom<HTMLDivElement>();
  const categoryOf = (entity: string, fault: string) => faultCategoryOf(plant, catalog.categories, entity, fault);
  return (
    <div
      ref={ref}
      className={cx(styles.scroll, shared.scrolls, shared.cut)}
      data-scroll="run-plan"
      data-cut={cut ? "true" : undefined}
    >
      {note}
      <PlanIdentity draft={draft} change={change} />
      <fieldset className={panel.group}>
        <legend>Run</legend>
        <Seed draft={draft} change={change} problems={shown} />
        <Duration draft={draft} change={change} problems={shown} />
      </fieldset>
      <Mode draft={draft} change={change} />
      <Pace draft={draft} change={change} durationS={plan.durationS} />
      <FaultRows
        plant={plant}
        draft={draft}
        problems={shown}
        categories={catalog.categories}
        selected={selected}
        change={change}
      />
      <DeliveryRows draft={draft} catalog={catalog.perturbations} problems={shown} change={change} />
      <CallRows plant={plant} draft={draft} problems={shown} selected={selected} change={change} />
      <PlanAgenda items={agendaOf(plan, categoryOf)} />
      <Readiness plan={plan} plantName={plantTitle(plant)} />
    </div>
  );
}

export function LaunchForm({ launch, plant, draft, selected, change, onStart, note }: FormProps) {
  const [tried, setTried] = useState(0);
  const [told, setTold] = useState("");
  const form = useFocusFirstProblem(tried);
  const catalog = usePlanCatalog(launch, plant.id);
  const plan = planOf(draft, plant, catalog.kinds);
  const problems = planProblems(plan, (entity, effector) => effectorOf(plant, entity, effector), catalog.kinds);
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
        shown={shown}
        plan={plan}
        note={note}
        catalog={catalog}
      />
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
