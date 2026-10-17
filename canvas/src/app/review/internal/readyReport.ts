import { runSpan } from "../../../model/review";
import type { Report } from "../../../model/review";
import type { ReviewSession } from "./reviewModule";
import type { ReviewState } from "./reviewStore";

export function readyReport(state: ReviewState): Report | null {
  return state.report.phase === "ready" ? state.report.value : null;
}

export function startNsOf(state: ReviewState, review: ReviewSession): number {
  return runSpan(readyReport(state), review.index.bounds()).startNs;
}
