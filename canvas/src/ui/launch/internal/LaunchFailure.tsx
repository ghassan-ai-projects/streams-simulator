import { panel, shared, cx, useFocusOnMount } from "../../primitives";
import type { LaunchState, LaunchStore } from "../../../app/launch";
import styles from "./LaunchPanel.module.css";

export function Failure({
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
        <strong>{state.kept === null ? "The run did not start." : "The run did not finish."}</strong> {state.words}
        {state.kept !== null && ` What ran before the failure is kept as run ${state.kept}.`}
      </p>
      <details className={styles.detail}>
        <summary>Binary message</summary>
        <p>{state.failure.diagnostic.message}</p>
      </details>
      <button
        type="button"
        ref={back}
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          launch.reset();
        }}
      >
        Change the plan
      </button>
    </div>
  );
}
