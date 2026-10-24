import type { ArgValue, LaunchPlan, SessionCheck, SessionDraft } from "../../../model/launch";
import type { Store } from "../../shared";
import type { ConnectState } from "./connectState";
import type { LaunchState } from "./launchState";
import type { PlanCatalogState } from "./planCatalog";
import type { PolicyDefaultState } from "./policyDefault";
import type { Proposal } from "./proposalSlot";
import type { LaunchOrigin } from "./trackedRun";

export type { ConnectState } from "./connectState";
export type { LaunchState } from "./launchState";
export type { PolicyDefaultState } from "./policyDefault";
export type { Loaded } from "./planCatalog";
export type { Proposal } from "./proposalSlot";

export type RunKind = "complete" | "session";

export interface ConnectSession extends Store<ConnectState> {
  readonly defaults: Store<PolicyDefaultState>;
  kindFor(domain: string): RunKind;
  chooseKind(domain: string, kind: RunKind): void;
  setUpNext(domain: string, policies: readonly string[]): void;
  setUpFresh(domain: string): void;
  unrecovered(domain: string): boolean;
  forgetUnrecovered(domain: string): void;
  draftFor(domain: string): SessionDraft;
  keepDraft(domain: string, draft: SessionDraft): void;
  prefill(domain: string): Promise<SessionDraft | null>;
  addPolicy(domain: string, digest: string): SessionDraft;
  saveDefault(domain: string): Promise<void>;
  check(domain: string, draft: SessionDraft): Promise<SessionCheck | null>;
  start(domain: string, draft: SessionDraft): Promise<void>;
  resume(): Promise<void>;
  stop(): Promise<void>;
  reset(): void;
}

export interface RunWatch {
  shows(state: LaunchState): state is Extract<LaunchState, { phase: "done" }>;
  holdFinished(then: (worldId: string) => void): () => void;
}

interface PlanCatalog extends Store<PlanCatalogState> {
  load(domain: string): void;
}

export interface LaunchStore extends Store<LaunchState> {
  readonly connect: ConnectSession;
  readonly catalog: PlanCatalog;
  propose(domain: string, plan: LaunchPlan, left?: readonly string[], from?: string): void;
  proposeFresh(domain: string, plan: LaunchPlan, from: string): void;
  forgetProposal(): void;
  proposal(domain: string): Proposal | null;
  nowMs(): number;
  watch(plant: string): RunWatch;
  start(origin: LaunchOrigin, plan: LaunchPlan): Promise<void>;
  resume(originOf: (domain: string) => LaunchOrigin | null): Promise<void>;
  pause(): Promise<void>;
  continueRun(): Promise<void>;
  step(): Promise<void>;
  setRate(rate: number): Promise<void>;
  invokeNow(entity: string, effector: string, args: Readonly<Record<string, ArgValue>>): Promise<void>;
  injectNow(entity: string, fault: string): Promise<void>;
  stopEarly(): void;
  reset(): void;
}
