import type { SyntheticEvent } from "react";
import { useEffect, useRef } from "react";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import type { PlanIssue, SessionDraft } from "../../../model/launch";
import { sessionStartWords } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { cx, holdEnter, panel, shared } from "../../primitives";
import { RealRows } from "./RealRows";
import { CheckResult } from "./SessionCheck";
import { SessionFields } from "./SessionFields";
import launch from "./LaunchPanel.module.css";

interface SessionFormProps {
  readonly plant: Plant;
  readonly draft: SessionDraft;
  readonly change: (next: SessionDraft) => void;
  readonly state: ConnectState;
  readonly connect: ConnectSession;
}

function liveIssues(state: ConnectState): readonly PlanIssue[] {
  return state.phase === "refused" && !state.stale ? state.issues : [];
}

function useFocusOnRefusal(state: ConnectState) {
  const form = useRef<HTMLFormElement>(null);
  const refused = state.phase === "refused" && !state.stale ? state : null;
  useEffect(() => {
    if (refused === null) return;
    const target = form.current?.querySelector<HTMLElement>("[aria-invalid='true'], [role='alert']");
    target?.focus();
  }, [refused]);
  return form;
}

export function SessionForm({ plant, draft, change, state, connect }: SessionFormProps) {
  const issues = liveIssues(state);
  const checking = state.phase === "checking";
  const form = useFocusOnRefusal(state);
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!checking) void connect.start(plant.id, draft);
  };
  return (
    <form
      ref={form}
      className={launch.plan}
      onSubmit={submit}
      onKeyDown={holdEnter}
      noValidate
      aria-label="Session plan"
    >
      <div className={cx(launch.scroll, shared.scrolls)}>
        <SessionFields draft={draft} change={change} issues={issues} />
        <RealRows plant={plant} draft={draft} change={change} issues={issues} />
        <CheckResult state={state} rows={draft.real.length} />
      </div>
      <div className={launch.footer}>
        <p id="session-summary" className={launch.summary}>
          {sessionStartWords}
        </p>
        <button
          type="button"
          className={cx(shared.secondary, panel.secondary)}
          aria-disabled={checking}
          onClick={() => {
            if (!checking) void connect.check(plant.id, draft);
          }}
        >
          Check the plan
        </button>
        <button type="submit" className={cx(shared.primary, panel.primary)} aria-describedby="session-summary">
          Start the session
        </button>
      </div>
    </form>
  );
}
