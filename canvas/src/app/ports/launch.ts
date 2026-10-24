import type { JobRequest, JobStatus, PlanIssue, SessionCheck } from "../../model/launch";
import type { RunFollower } from "./review";
import type { Diagnostic, Outcome } from "./shared";

export type { JobRequest, JobStatus, PlannedInstance, SessionCheck } from "../../model/launch";

export type LaunchFailure =
  | {
      readonly kind: "rejected";
      readonly step: "start" | "status" | "stop" | "list" | "run";
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
  startJob(request: JobRequest): Promise<Outcome<JobStatus, LaunchFailure>>;
  jobStatus(worldId: string): Promise<Outcome<JobStatus, LaunchFailure>>;
  stopJob(worldId: string): Promise<Outcome<JobStatus, LaunchFailure>>;
  listJobs(): Promise<Outcome<readonly JobStatus[], LaunchFailure>>;
}
