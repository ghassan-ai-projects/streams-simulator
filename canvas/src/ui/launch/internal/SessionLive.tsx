import type { ConnectSession, ConnectState } from "../../../app/launch";
import type { BindingNote } from "../../../model/launch";
import { bindingWords } from "../../../model/launch";
import { cx, failureLine, panel, shared, useFocusOnMount } from "../../primitives";
import { IssueList, Planned } from "./SessionCheck";
import launch from "./LaunchPanel.module.css";

type Live = Extract<ConnectState, { phase: "live" | "starting" }>;
type Failed = Extract<ConnectState, { phase: "failed" }>;

function liveWords(state: Live): string {
  if (state.phase === "starting") return "Starting the session…";
  return state.stopping ? "Stopping and saving the session…" : `Connected session · run ${state.run}`;
}

function BindingResult({ binding }: { readonly binding: BindingNote }) {
  return (
    <div className={binding.matched ? undefined : launch.failure} role={binding.matched ? "status" : "alert"}>
      <p>{bindingWords(binding)}</p>
      <IssueList label="Why the gateway was refused" issues={binding.issues} />
    </div>
  );
}

function Binding({ state }: { readonly state: Live }) {
  if (state.phase !== "live") return null;
  if (state.binding !== null) return <BindingResult binding={state.binding} />;
  if (state.check === null) return null;
  return (
    <p className={panel.hint}>
      {state.check.mixed
        ? "Waiting for the gateway to announce its boards…"
        : "No gateway: every instance is simulated."}
    </p>
  );
}

export function SessionLive({ state, connect }: { readonly state: Live; readonly connect: ConnectSession }) {
  const status = useFocusOnMount<HTMLParagraphElement>();
  const stopping = state.phase === "starting" || state.stopping;
  return (
    <div className={panel.body}>
      <p className={launch.status} role="status" ref={status} tabIndex={-1}>
        {liveWords(state)}
      </p>
      <Binding state={state} />
      {state.check !== null && <Planned check={state.check} />}
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        aria-disabled={stopping}
        onClick={() => {
          if (!stopping) void connect.stop();
        }}
      >
        Stop and save
      </button>
      <p className={panel.hint}>
        The session runs until you stop it, also when you close this panel or leave the plant. Stop ends it and saves
        the run folder, which you can review and replay.
      </p>
    </div>
  );
}

export function SessionEnded({ run, onOpen }: { readonly run: string; readonly onOpen: () => void }) {
  const status = useFocusOnMount<HTMLParagraphElement>();
  return (
    <div className={panel.body}>
      <p className={launch.status} role="status" ref={status} tabIndex={-1}>
        {`The session was stopped and saved as run ${run}.`}
      </p>
      <button type="button" className={cx(shared.primary, panel.primary)} onClick={onOpen}>
        Open the review
      </button>
    </div>
  );
}

export function SessionFailed({ state, onBack }: { readonly state: Failed; readonly onBack: () => void }) {
  return (
    <div className={cx(panel.body, launch.failure)}>
      <p role="alert">{failureLine(state.diagnostic)}</p>
      <IssueList label="Problems the session reported" issues={state.issues} />
      {state.run !== null && <p className={panel.hint}>{`What ran is kept in run ${state.run}.`}</p>}
      <button type="button" className={cx(shared.secondary, panel.secondary)} onClick={onBack}>
        Back to the session plan
      </button>
    </div>
  );
}

export function SessionElsewhere({ domain }: { readonly domain: string }) {
  return (
    <div className={panel.body}>
      <p className={launch.status} role="status">{`A connected session of ${domain} is running.`}</p>
      <p className={panel.hint}>One session at a time: stop it on its plant before starting another.</p>
    </div>
  );
}
