import type { Clock, ReviewGateway, Scheduler } from "../../ports";
import { createOperationIdentity } from "../../shared";
import type { ReviewSession } from "./api";
import { ReviewStore } from "./reviewStore";

export function createReviewSession(gateway: ReviewGateway, clock: Clock, scheduler: Scheduler): ReviewSession {
  return new ReviewStore(gateway, clock, scheduler, createOperationIdentity());
}
