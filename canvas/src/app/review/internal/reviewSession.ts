import type { Clock, ReviewGateway, Scheduler } from "../../ports";
import { createOperationIdentity } from "../../shared";
import type { ReviewSession } from "./api";
import type { ControlSession } from "./controlStore";
import { ControlStore } from "./controlStore";
import { ReviewStore } from "./reviewStore";
import type { TwinSession } from "./twinStore";
import { TwinStore } from "./twinStore";

class ControlledReview extends ReviewStore implements ReviewSession {
  readonly controls: ControlSession;
  readonly twin: TwinSession;

  constructor(gateway: ReviewGateway, clock: Clock, scheduler: Scheduler) {
    super(gateway, clock, scheduler, createOperationIdentity());
    this.controls = new ControlStore(gateway, {
      run: () => this.snapshot().worldId,
      newest: () => this.snapshot().liveHistory?.entries.at(-1) ?? null,
    });
    this.subscribe(() => {
      this.controls.bind(this.snapshot().worldId);
    });
    this.twin = new TwinStore(gateway);
  }
}

export function createReviewSession(gateway: ReviewGateway, clock: Clock, scheduler: Scheduler): ReviewSession {
  return new ControlledReview(gateway, clock, scheduler);
}
