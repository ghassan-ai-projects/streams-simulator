import type { BindingNote, PlanIssue, SessionCheck } from "../../../model/launch";
import type { ConnectFailure, ConnectStatus, Diagnostic } from "../../ports";
export type ConnectState =
  | { readonly phase: "idle" }
  | { readonly phase: "checking"; readonly domain: string }
  | { readonly phase: "checked"; readonly domain: string; readonly check: SessionCheck; readonly stale: boolean }
  | {
      readonly phase: "refused";
      readonly domain: string;
      readonly diagnostic: Diagnostic;
      readonly issues: readonly PlanIssue[];
      readonly stale: boolean;
    }
  | { readonly phase: "starting"; readonly domain: string; readonly check: SessionCheck | null }
  | {
      readonly phase: "live";
      readonly domain: string;
      readonly run: string;
      readonly check: SessionCheck | null;
      readonly binding: BindingNote | null;
      readonly stopping: boolean;
      readonly controls: boolean;
      readonly note: string | null;
    }
  | { readonly phase: "ended"; readonly domain: string; readonly run: string }
  | {
      readonly phase: "failed";
      readonly domain: string;
      readonly run: string | null;
      readonly diagnostic: Diagnostic;
      readonly issues: readonly PlanIssue[];
    };

export type LiveSession = Extract<ConnectState, { phase: "live" }>;

const busyPhases: readonly ConnectState["phase"][] = ["checking", "starting", "live"];

export function isBusy(state: ConnectState): boolean {
  return busyPhases.includes(state.phase);
}

export function refusedState(domain: string, failure: ConnectFailure): ConnectState {
  if (failure.kind === "refused") {
    return { phase: "refused", domain, diagnostic: failure.diagnostic, issues: failure.issues, stale: false };
  }
  return { phase: "failed", domain, run: null, diagnostic: failure.diagnostic, issues: [] };
}

interface LiveFlags {
  readonly stopping: boolean;
  readonly controls: boolean;
}

function liveState(domain: string, run: string, check: SessionCheck | null, flags: LiveFlags): LiveSession {
  return { phase: "live", domain, run, check, binding: null, ...flags, note: null };
}

export function stateOf(domain: string, check: SessionCheck | null, status: ConnectStatus): ConnectState {
  if (status.state === "none") return { phase: "idle" };
  const own = status.domain === "" ? domain : status.domain;
  const plan = status.check ?? check;
  switch (status.state) {
    case "starting":
      return { phase: "starting", domain: own, check: plan };
    case "running":
    case "stopping":
      return liveState(own, status.run, plan, { stopping: status.state === "stopping", controls: status.controls });
    case "ended":
      return { phase: "ended", domain: own, run: status.run };
    case "failed":
      return { phase: "failed", domain: own, run: status.run, diagnostic: status.diagnostic, issues: status.issues };
  }
}

export function markedStale(state: ConnectState, domain: string): ConnectState {
  if ((state.phase === "checked" || state.phase === "refused") && state.domain === domain && !state.stale) {
    return { ...state, stale: true };
  }
  return state;
}
