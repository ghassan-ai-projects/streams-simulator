import type { LaunchPlan } from "../../../model/launch";
import { readinessOf } from "../../../model/launch";
import { cx, panel } from "../../primitives";
import styles from "./Readiness.module.css";

export function Readiness({ plan, plantName }: { readonly plan: LaunchPlan; readonly plantName: string }) {
  return (
    <section className={panel.group} aria-labelledby="launch-readiness-title">
      <h3 id="launch-readiness-title">Before you start</h3>
      <dl className={styles.lines}>
        {readinessOf(plan, plantName).map((line) => (
          <div key={line.label} className={styles.line}>
            <dt>{line.label}</dt>
            <dd className={cx(line.label === "Scoring" && styles.scoring)}>{line.text}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
