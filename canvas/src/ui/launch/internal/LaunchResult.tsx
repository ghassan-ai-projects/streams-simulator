import { cx, panel, shared } from "../../primitives";
import type { LaunchState } from "../../../app/launch";
import { outcomeWords, planSummary } from "../../../model/launch";
import { clockOf, durationOf } from "../../../model/shared/duration";
import { useFocusOnMount } from "../../primitives";
import styles from "./LaunchPanel.module.css";

type Done = Extract<LaunchState, { phase: "done" }>;

function callsWords(state: Done): string {
  const refused = state.calls.filter((call) => !call.accepted).length;
  if (state.calls.length === 0) return "";
  return refused === 0
    ? " Every effector call was accepted."
    : ` ${String(refused)} of ${String(state.calls.length)} effector calls were not accepted.`;
}

function resultWords(state: Done): string {
  const summary = planSummary(state.plan, durationOf(state.plan.durationS));
  if (!state.early) return `${summary}.${callsWords(state)}`;
  const reached = `${clockOf(state.reachedS)} of ${clockOf(state.plan.durationS)}`;
  return `Saved at ${reached} as a shorter run · ${summary}.${callsWords(state)}`;
}

export function Result({ state, onOpen }: { readonly state: Done; readonly onOpen: () => void }) {
  const status = useFocusOnMount<HTMLParagraphElement>();
  return (
    <div className={panel.body}>
      <p className={styles.status} role="status" ref={status} tabIndex={-1}>
        {resultWords(state)}
      </p>
      {state.calls.length > 0 && (
        <ul className={styles.outcomes} aria-label="Effector call results">
          {state.calls.map((outcome, index) => (
            <li key={String(index)}>{outcomeWords(outcome)}</li>
          ))}
        </ul>
      )}
      <button type="button" className={cx(shared.primary, panel.primary)} onClick={onOpen}>
        Open the review
      </button>
    </div>
  );
}
