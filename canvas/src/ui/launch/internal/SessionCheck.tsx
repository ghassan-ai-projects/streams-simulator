import { CircuitBoard, SquareDashed, TriangleAlert } from "lucide-react";
import type { ConnectState } from "../../../app/launch";
import type { InstanceRole, PlanIssue, SessionCheck } from "../../../model/launch";
import { checkWords, plannedWords, roleWords, unplacedIssues } from "../../../model/launch";
import { cx, failureLine, panel } from "../../primitives";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

const icons = { real: CircuitBoard, simulated: SquareDashed } as const;

function RoleMark({ role }: { readonly role: InstanceRole }) {
  const Icon = icons[role];
  return (
    <span className={cx(styles.role, styles[role])} title={roleWords[role]}>
      <Icon size={13} aria-hidden="true" />
    </span>
  );
}

export function Planned({ check }: { readonly check: SessionCheck }) {
  return (
    <section className={panel.group} aria-label="Instances in this session">
      <p className={launch.status} role="status">
        {checkWords(check)}
      </p>
      <ul className={styles.instances}>
        {check.instances.map((planned) => (
          <li key={planned.instance}>
            <RoleMark role={planned.role} />
            {plannedWords(planned)}
          </li>
        ))}
      </ul>
    </section>
  );
}

export function IssueList({ label, issues }: { readonly label: string; readonly issues: readonly PlanIssue[] }) {
  if (issues.length === 0) return null;
  return (
    <ul className={styles.issues} aria-label={label}>
      {issues.map((issue) => (
        <li key={`${issue.pointer}:${issue.message}`} className={launch.problem}>
          <TriangleAlert size={13} aria-hidden="true" />
          {issue.message}
        </li>
      ))}
    </ul>
  );
}

export function CheckResult({ state }: { readonly state: ConnectState }) {
  if (state.phase === "checking") {
    return (
      <p className={panel.hint} role="status">
        Checking the plan with streamsim…
      </p>
    );
  }
  if (state.phase === "refused") {
    return (
      <div className={launch.failure} role="alert">
        <p>{failureLine(state.diagnostic)}</p>
        <IssueList label="Other problems in the plan" issues={unplacedIssues(state.issues)} />
      </div>
    );
  }
  return state.phase === "checked" ? <Planned check={state.check} /> : null;
}
