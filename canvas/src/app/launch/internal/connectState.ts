import type { BindingNote, PlanIssue, SessionCheck } from "../../../model/launch";
import type { ConnectFailure, ConnectStatus, Diagnostic } from "../../ports";

export type ConnectState =
  | { readonly phase: "idle" }
  | { readonly phase: "checking"; readonly domain: string }
  | { readonly phase: "checked"; readonly domain: string; readonly check: SessionCheck }
  | {
      readonly phase: "refused";
      readonly domain: string;
      readonly diagnostic: Diagnostic;
      readonly issues: readonly PlanIssue[];
    }
  | { readonly phase: "starting"; readonly domain: string; readonly check: SessionCheck }
  | {
      readonly phase: "live";
      readonly domain: string;
      readonly run: string;
      readonly check: SessionCheck | null;
      readonly binding: BindingNote | null;
      readonly stopping: boolean;
    }
  | { readonly phase: "ended"; readonly domain: string; readonly run: string }
  | {
      readonly phase: "failed";
      readonly domain: string;
      readonly run: string | null;
      readonly diagnostic: Diagnostic;
      readonly issues: readonly PlanIssue[];
    };

export type Live = Extract<ConnectState, { phase: "live" }>;

const busyPhases: readonly ConnectState["phase"][] = ["checking", "starting", "live"];

export function isBusy(state: ConnectState): boolean {
  return busyPhases.includes(state.phase);
}

export function refusedState(domain: string, failure: ConnectFailure): ConnectState {
  if (failure.kind === "refused") {
    return { phase: "refused", domain, diagnostic: failure.diagnostic, issues: failure.issues };
  }
  return { phase: "failed", domain, run: null, diagnostic: failure.diagnostic, issues: [] };
}

export function stateOf(domain: string, check: SessionCheck | null, status: ConnectStatus): ConnectState {
  switch (status.state) {
    case "running":
      return { phase: "live", domain, run: status.run, check, binding: null, stopping: false };
    case "ended":
      return { phase: "ended", domain, run: status.run };
    case "failed":
      return { phase: "failed", domain, run: status.run, diagnostic: status.diagnostic, issues: status.issues };
    case "none":
      return { phase: "idle" };
  }
}
