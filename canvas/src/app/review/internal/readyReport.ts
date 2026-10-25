import { runSpan } from "../../../model/review";
import type { PlaybackIndex, Report } from "../../../model/review";
import type { ReviewState } from "./reviewState";

export { runGoing } from "./reviewState";

export function readyReport(state: ReviewState): Report | null {
  return state.report.phase === "ready" ? state.report.value : null;
}

export function startNsOf(state: ReviewState, review: { readonly index: PlaybackIndex }): number {
  const report = readyReport(state);
  if (report === null && review.index.openedNs > 0) return review.index.openedNs;
  return runSpan(report, review.index.bounds()).startNs;
}
