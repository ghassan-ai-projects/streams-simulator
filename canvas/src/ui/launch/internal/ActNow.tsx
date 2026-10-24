import type { LaunchStore } from "../../../app/launch";
import { exploreOnlyFaultsWords, nowLines } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { panel } from "../../primitives";
import { CallNow } from "./CallNow";
import { FaultNow } from "./FaultNow";
import type { Running } from "./RunControls";
import styles from "./RunControls.module.css";

function NowResults({ state }: { readonly state: Running }) {
  const lines = nowLines(state.calls, state.injected);
  if (lines.length === 0) return null;
  return (
    <ul className={styles.results} aria-label="Commands given now">
      {lines.map((words, index) => (
        <li key={String(index)}>{words}</li>
      ))}
    </ul>
  );
}

interface ActNowProps {
  readonly state: Running;
  readonly launch: LaunchStore;
  readonly plant: Plant;
}

export function ActNow({ state, launch, plant }: ActNowProps) {
  return (
    <fieldset className={panel.group}>
      <legend>Act now</legend>
      <CallNow plant={plant} launch={launch} />
      {state.plan.explore ? (
        <FaultNow plant={plant} launch={launch} state={state} />
      ) : (
        <p className={panel.hint}>{exploreOnlyFaultsWords}</p>
      )}
      <NowResults state={state} />
      <p className={panel.hint}>
        A command given now applies at the simulated instant the run has reached and is recorded in the run.
      </p>
    </fieldset>
  );
}
