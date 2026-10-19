import type { Badge, PlaybackIndex } from "../../../model/review";
import type { RunEntry } from "../../../model/workspace";
import type { Store } from "../../shared";
import type { ControlSession } from "./controlStore";
import type { ReviewState } from "./reviewState";

export type { ReviewState } from "./reviewState";
export type { ReviewPlant } from "./reviewPlant";
export type { ControlSession } from "./controlStore";

export interface ReviewSession extends Store<ReviewState> {
  readonly index: PlaybackIndex;
  readonly controls: ControlSession;
  open(worldId: string, known?: RunEntry): void;
  openLive(worldId: string, known?: RunEntry): void;
  refreshCompleted(worldId: string, known: RunEntry | undefined): void;
  close(): void;
  badgeFor(worldId: string): Badge | null;
  seek(positionNs: number): void;
  backToLive(): void;
  play(): void;
  pause(): void;
  setSpeed(speed: number): void;
  select(entity: string | null): void;
  reveal(positionNs: number, entity: string): void;
  showTab(tab: ReviewState["tab"]): void;
}
