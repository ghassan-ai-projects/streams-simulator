import type { Acknowledgement, PlanIssue, SessionCheck } from "../../model/launch";
import type { RunFollower } from "./review";
import type { Diagnostic, Outcome } from "./shared";

export type { Acknowledgement, PlannedInstance, SessionCheck } from "../../model/launch";

export type LaunchFailure =
  | {
      readonly kind: "rejected";
      readonly step: "create" | "advance" | "inject" | "invoke" | "end";
      readonly diagnostic: Diagnostic;
    }
  | { readonly kind: "unavailable" | "unreadable"; readonly diagnostic: Diagnostic };

export type ConnectFailure =
  | { readonly kind: "refused"; readonly diagnostic: Diagnostic; readonly issues: readonly PlanIssue[] }
  | { readonly kind: "busy" | "unavailable" | "unreadable"; readonly diagnostic: Diagnostic };

interface SessionFacts {
  readonly run: string;
  readonly domain: string;
  readonly check: SessionCheck | null;
  readonly controls: boolean;
}

export type ConnectStatus =
  | { readonly state: "none" }
  | (SessionFacts & { readonly state: "starting" | "running" | "stopping" | "ended" })
  | (SessionFacts & {
      readonly state: "failed";
      readonly diagnostic: Diagnostic;
      readonly issues: readonly PlanIssue[];
    });

export type PolicyDefaultFile =
  | { readonly state: "absent" }
  | { readonly state: "present"; readonly policies: readonly string[]; readonly etag: string }
  | { readonly state: "unreadable"; readonly etag: string };

export interface PolicyDefaultFailure {
  readonly kind: "conflict" | "unavailable" | "unreadable";
  readonly diagnostic: Diagnostic;
}

export interface ConnectGateway {
  readPolicyDefault(): Promise<Outcome<PolicyDefaultFile, PolicyDefaultFailure>>;
  savePolicyDefault(
    policies: readonly string[],
    etag: string,
  ): Promise<Outcome<{ readonly etag: string }, PolicyDefaultFailure>>;
  checkSession(plan: string): Promise<Outcome<SessionCheck, ConnectFailure>>;
  startSession(plan: string): Promise<Outcome<ConnectStatus, ConnectFailure>>;
  sessionStatus(): Promise<Outcome<ConnectStatus, ConnectFailure>>;
  stopSession(): Promise<Outcome<ConnectStatus, ConnectFailure>>;
  followSession(run: string, follower: RunFollower): () => void;
}

export interface LaunchGateway extends ConnectGateway {
  createWorld(input: {
    readonly domain: string;
    readonly seed: number;
    readonly explore: boolean;
  }): Promise<Outcome<{ readonly worldId: string }, LaunchFailure>>;
  advance(worldId: string, byNs: number): Promise<Outcome<void, LaunchFailure>>;
  inject(worldId: string, entity: string, fault: string): Promise<Outcome<void, LaunchFailure>>;
  invoke(
    worldId: string,
    entity: string,
    effector: string,
    args: Readonly<Record<string, unknown>>,
  ): Promise<Outcome<Acknowledgement, LaunchFailure>>;
  endRun(worldId: string): Promise<Outcome<void, LaunchFailure>>;
  destroyWorld(worldId: string): Promise<void>;
}
