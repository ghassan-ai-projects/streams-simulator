import { TriangleAlert } from "lucide-react";
import { useEffect, useRef } from "react";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import type { BindingNote, PlanIssue, PlannedInstance, SessionCheck } from "../../../model/launch";
import {
  checkWords,
  isPolicyIssue,
  issueDetail,
  issueWords,
  plannedWords,
  refusedFieldWords,
  refusedPlanWords,
  staleCheckWords,
} from "../../../model/launch";
import { originLabels } from "../../../model/shared/origin";
import { failureLine, OriginBadge, panel } from "../../primitives";
import launch from "./LaunchPanel.module.css";
import { PolicyInEffect, SaveDefault } from "./SessionPolicies";
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

function Developers({ issues, raw }: { readonly issues: readonly PlanIssue[]; readonly raw: string }) {
  if (issues.length === 0 && raw === "") return null;
  return (
    <details className={styles.developers}>
      <summary>For developers</summary>
      {raw !== "" && <p>{raw}</p>}
      <ul className={styles.issues}>
        {issues.map((issue, index) => (
          <li key={`${String(index)}:${issue.pointer}`}>{issueDetail(issue)}</li>
        ))}
      </ul>
    </details>
  );
}

interface IssueListProps {
  readonly label: string;
  readonly issues: readonly PlanIssue[];
  readonly raw?: string;
  readonly marked?: readonly PlanIssue[];
}

export function IssueList({ label, issues, raw = "", marked = [] }: IssueListProps) {
  const listed = issues.filter((issue) => !marked.includes(issue));
  return (
    <>
      {listed.length > 0 && (
        <ul className={styles.issues} aria-label={label}>
          {listed.map((issue, index) => (
            <li key={`${String(index)}:${issue.pointer}`} className={launch.problem}>
              <TriangleAlert size={13} aria-hidden="true" />
              {issueWords(issue)}
            </li>
          ))}
        </ul>
      )}
      <Developers issues={issues} raw={raw} />
    </>
  );
}

interface CheckResultProps {
  readonly state: ConnectState;
  readonly connect: ConnectSession;
  readonly domain: string;
}

type CheckedState = Extract<ConnectState, { phase: "checked" }>;

function Refused({ state }: { readonly state: Extract<ConnectState, { phase: "refused" }> }) {
  if (state.stale) return <p className={panel.hint}>{staleCheckWords}</p>;
  const onField = state.issues.filter(isPolicyIssue);
  const onlyField = onField.length > 0 && onField.length === state.issues.length;
  return (
    <div className={launch.failure} role="alert" tabIndex={-1}>
      <p>{onlyField ? refusedFieldWords : refusedPlanWords}</p>
      <IssueList
        label="Problems in the plan"
        issues={state.issues}
        raw={failureLine(state.diagnostic)}
        marked={onField}
      />
    </div>
  );
}

interface CheckedProps {
  readonly state: CheckedState;
  readonly connect: ConnectSession;
  readonly domain: string;
}

function Checked({ state, connect, domain }: CheckedProps) {
  const result = useRef<HTMLElement>(null);
  const fresh = state.stale ? null : state;
  useEffect(() => {
    if (fresh !== null) result.current?.focus();
  }, [fresh]);
  return (
    <>
      {state.stale && <p className={panel.hint}>{staleCheckWords}</p>}
      <section ref={result} className={styles.checked} aria-label="Plan check result" tabIndex={-1}>
        <Planned check={state.check} role="status" />
        <PolicyInEffect policies={state.check.policies} />
        {!state.stale && <SaveDefault connect={connect} domain={domain} />}
      </section>
    </>
  );
}

export function CheckResult({ state, connect, domain }: CheckResultProps) {
  if (state.phase === "checking") {
    return (
      <p className={panel.hint} role="status">
        Checking the plan with streamsim…
      </p>
    );
  }
  if (state.phase === "refused") return <Refused state={state} />;
  if (state.phase !== "checked") return null;
  return <Checked state={state} connect={connect} domain={domain} />;
}
