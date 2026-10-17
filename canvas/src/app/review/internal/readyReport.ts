import { runSpan } from "../../../model/review";
import type { PlaybackIndex, Report } from "../../../model/review";
import type { ReviewState } from "./reviewState";

export function readyReport(state: ReviewState): Report | null {
  return state.report.phase === "ready" ? state.report.value : null;
}

export function startNsOf(state: ReviewState, review: { readonly index: PlaybackIndex }): number {
  return runSpan(readyReport(state), review.index.bounds()).startNs;
}
