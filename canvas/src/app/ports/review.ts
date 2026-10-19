import type {
  ControlFailureKind,
  ControlReply,
  ControlRequest,
  JournalEntry,
  TwinReport,
  TwinRequest,
} from "../../model/review";
import type { Report, Replay } from "../../model/review";
import type { Diagnostic, DomainDocument, Outcome } from "./shared";

export type ReviewFailure =
  | { readonly kind: "missing" | "run-incomplete"; readonly diagnostic: Diagnostic }
  | { readonly kind: "unavailable" | "unreadable"; readonly diagnostic: Diagnostic };

export interface ControlFailure {
  readonly kind: ControlFailureKind;
  readonly diagnostic: Diagnostic;
}

export interface TwinFailure {
  readonly kind: "refused" | "unavailable" | "unreadable";
  readonly diagnostic: Diagnostic;
}

export interface RunFollower {
  entry(entry: JournalEntry): void;
  caughtUp?(): void;
  closed(): void;
  failed(): void;
}

export interface ReviewGateway {
  report(runId: string): Promise<Outcome<Report, ReviewFailure>>;
  verify(runId: string): Promise<Outcome<Replay, ReviewFailure>>;
  runDomain(runId: string): Promise<Outcome<DomainDocument, ReviewFailure>>;
  follow(runId: string, follower: RunFollower): () => void;
  control(request: ControlRequest): Promise<Outcome<ControlReply, ControlFailure>>;
  twinReport(request: TwinRequest): Promise<Outcome<TwinReport, TwinFailure>>;
}
