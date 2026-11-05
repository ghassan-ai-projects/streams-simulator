import { runSpan } from "../../../model/review";
import type { PlaybackIndex, Report } from "../../../model/review";
import type { ReviewState } from "./reviewState";

export { runGoing } from "./reviewState";

export function readyReport(state: ReviewState): Report | null {
  return state.report.phase === "ready" ? state.report.value : null;
}

export function timelineSpan(
  state: ReviewState,
  review: { readonly index: PlaybackIndex },
): { readonly startNs: number; readonly endNs: number; readonly positionNs: number } {
  const span = runSpan(readyReport(state), review.index.bounds());
  return {
    startNs: Math.min(span.startNs, state.liveHistory?.startNs ?? span.startNs),
    endNs: Math.max(span.endNs, state.newestNs),
    positionNs: state.positionNs === 0 ? span.startNs : state.positionNs,
  };
}

export function startNsOf(state: ReviewState, review: { readonly index: PlaybackIndex }): number {
  const report = readyReport(state);
  if (report === null && review.index.openedNs > 0) return review.index.openedNs;
  return runSpan(report, review.index.bounds()).startNs;
}
