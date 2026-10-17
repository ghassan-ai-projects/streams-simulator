import type { Badge, PlaybackIndex } from "../../../model/review";
import type { RunEntry } from "../../../model/workspace";
import type { Clock, ReviewGateway, Scheduler } from "../../ports";
import { createOperationIdentity } from "../../shared";
import { ReviewStore } from "./reviewStore";
import type { ReviewState } from "./reviewStore";

export interface ReviewSession {
  readonly snapshot: () => ReviewState;
  readonly subscribe: (listener: () => void) => () => void;
  readonly index: PlaybackIndex;
  open(worldId: string, known?: RunEntry): void;
  refreshCompleted(worldId: string, known: RunEntry | undefined): void;
  close(): void;
  badgeFor(worldId: string): Badge | null;
  seek(positionNs: number): void;
  play(): void;
  pause(): void;
  setSpeed(speed: number): void;
  tick(elapsedMs: number): void;
  select(entity: string | null): void;
  reveal(positionNs: number, entity: string): void;
  showTab(tab: ReviewState["tab"]): void;
}

export interface ReviewModule {
  createSession(): ReviewSession;
}

export function createReviewModule(gateway: ReviewGateway, clock: Clock, scheduler: Scheduler): ReviewModule {
  return {
    createSession: () => new ReviewStore(gateway, clock, scheduler, createOperationIdentity()),
  };
}
