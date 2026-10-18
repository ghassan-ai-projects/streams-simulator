import type { LaunchPlan, SessionCheck, SessionDraft } from "../../../model/launch";
import type { Store } from "../../shared";
import type { ConnectState } from "./connectState";
import type { LaunchState } from "./launchState";
import type { Proposal } from "./proposalSlot";
import type { LaunchOrigin } from "./runCursor";

export type { ConnectState } from "./connectState";
export type { LaunchState } from "./launchState";
export type { Proposal } from "./proposalSlot";

export interface ConnectSession extends Store<ConnectState> {
  check(domain: string, draft: SessionDraft): Promise<SessionCheck | null>;
  start(domain: string, draft: SessionDraft): Promise<void>;
  resume(domain: string): Promise<void>;
  stop(): Promise<void>;
  reset(): void;
}

export interface RunWatch {
  shows(state: LaunchState): state is Extract<LaunchState, { phase: "done" }>;
  holdFinished(then: (worldId: string) => void): () => void;
}

export interface LaunchStore extends Store<LaunchState> {
  readonly connect: ConnectSession;
  propose(domain: string, plan: LaunchPlan, left?: readonly string[], from?: string): void;
  forgetProposal(): void;
  proposal(domain: string): Proposal | null;
  nowMs(): number;
  watch(plant: string): RunWatch;
  start(origin: LaunchOrigin, plan: LaunchPlan): Promise<void>;
  stopEarly(): void;
  reset(): void;
}
