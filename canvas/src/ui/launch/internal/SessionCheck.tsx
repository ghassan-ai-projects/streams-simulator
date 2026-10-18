import { TriangleAlert } from "lucide-react";
import type { ConnectState } from "../../../app/launch";
import type { BindingNote, PlanIssue, PlannedInstance, SessionCheck } from "../../../model/launch";
import {
  checkWords,
  issueDetail,
  issueWords,
  plannedWords,
  staleCheckWords,
  unplacedIssues,
} from "../../../model/launch";
import { originLabels } from "../../../model/shared/origin";
import { failureLine, OriginBadge, panel } from "../../primitives";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

function notBound(planned: PlannedInstance, binding: BindingNote | null): boolean {
  return planned.role === "real" && binding !== null && !binding.matched;
}

function PlannedRow({ planned, binding }: { readonly planned: PlannedInstance; readonly binding: BindingNote | null }) {
  const words = plannedWords(planned, binding);
  return (
    <li className={notBound(planned, binding) ? launch.problem : undefined}>
      {notBound(planned, binding) ? (
        <TriangleAlert size={13} aria-hidden="true" />
      ) : (
        <OriginBadge kind={planned.role} label="" title={originLabels[planned.role]} said="" variant="glyph" />
      )}
      {words}
    </li>
  );
}

interface PlannedProps {
  readonly check: SessionCheck;
  readonly binding?: BindingNote | null;
  readonly role?: "status";
}

export function Planned({ check, binding = null, role }: PlannedProps) {
  return (
    <section className={panel.group} aria-label="Instances in this session">
      <p className={launch.status} role={role}>
        {checkWords(check)}
      </p>
      <ul className={styles.instances}>
        {check.instances.map((planned) => (
          <PlannedRow key={planned.instance} planned={planned} binding={binding} />
        ))}
      </ul>
    </section>
  );
}

export function IssueList({ label, issues }: { readonly label: string; readonly issues: readonly PlanIssue[] }) {
  if (issues.length === 0) return null;
  return (
    <>
      <ul className={styles.issues} aria-label={label}>
        {issues.map((issue, index) => (
          <li key={`${String(index)}:${issue.pointer}`} className={launch.problem}>
            <TriangleAlert size={13} aria-hidden="true" />
            {issueWords(issue)}
          </li>
        ))}
      </ul>
      <details className={styles.developers}>
        <summary>For developers</summary>
        <ul className={styles.issues}>
          {issues.map((issue, index) => (
            <li key={`${String(index)}:${issue.pointer}`}>{issueDetail(issue)}</li>
          ))}
        </ul>
      </details>
    </>
  );
}

function Refused({
  state,
  rows,
}: {
  readonly state: Extract<ConnectState, { phase: "refused" }>;
  readonly rows: number;
}) {
  return (
    <div className={launch.failure} role="alert" tabIndex={-1}>
      <p>{failureLine(state.diagnostic)}</p>
      {state.stale ? (
        <p>{staleCheckWords}</p>
      ) : (
        <IssueList label="Other problems in the plan" issues={unplacedIssues(state.issues, rows)} />
      )}
    </div>
  );
}

export function CheckResult({ state, rows }: { readonly state: ConnectState; readonly rows: number }) {
  if (state.phase === "checking") {
    return (
      <p className={panel.hint} role="status">
        Checking the plan with streamsim…
      </p>
    );
  }
  if (state.phase === "refused") return <Refused state={state} rows={rows} />;
  if (state.phase !== "checked") return null;
  return (
    <>
      {state.stale && <p className={panel.hint}>{staleCheckWords}</p>}
      <Planned check={state.check} role="status" />
    </>
  );
}
