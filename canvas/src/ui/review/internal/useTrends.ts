import { useMemo } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { readyReport, runGoing, startNsOf } from "../../../app/review";
import { overlayOf, plotsOf, runSpan, trendSpanOf } from "../../../model/review";
import type { OverlayEvent, PlannedFault, Plot, Span, TrendSpan, Watched, WindowChoice } from "../../../model/review";
import { useStore } from "../../primitives";

const nothing: readonly Watched[] = [];

export interface TrendsView {
  readonly open: boolean;
  readonly watched: readonly Watched[];
  readonly choice: WindowChoice;
  readonly span: Span;
  readonly limitedFromNs: number | null;
  readonly running: boolean;
  readonly startNs: number;
  readonly cursorNs: number;
  readonly plots: readonly Plot[];
  readonly events: readonly OverlayEvent[];
}

function useWindow(review: ReviewSession, state: ReviewState, watched: readonly Watched[]) {
  const choice = useStore(review.watchList).choice;
  const running = runGoing(state);
  const reachedNs = Math.max(state.newestNs, review.index.clockNs, review.index.bounds()?.endNs ?? 0);
  const startNs = startNsOf(state, review);
  const report = readyReport(state);
  const trend: TrendSpan = trendSpanOf(review.index, watched, {
    choice,
    running,
    reachedNs,
    cursorNs: state.positionNs,
    runStartNs: startNs,
    runEndNs: runSpan(report, review.index.bounds()).endNs,
  });
  const { startNs: from, endNs: to } = trend.span;
  const span = useMemo(() => ({ startNs: from, endNs: to }), [from, to]);
  const live = running && state.following ? reachedNs : state.positionNs;
  const cursorNs = Math.min(to, Math.max(from, live));
  return { trend, span, running, startNs, cursorNs, report };
}

export function useTrends(
  review: ReviewSession,
  state: ReviewState,
  plant: ReadyPlant | null,
  width: number,
  faults: readonly PlannedFault[],
): TrendsView {
  const watch = useStore(review.watchList);
  const watched = watch.lists.get(plant?.plant.id ?? "") ?? nothing;
  const { trend, span, running, startNs, cursorNs, report } = useWindow(review, state, watched);
  const plots = plotsOf(review.index, watched, span, width);
  const events = overlayOf(review.index.calls(span.startNs, span.endNs), report, span, { startNs, faults });
  const { choice, limitedFromNs } = trend;
  return { open: watch.open, watched, choice, span, limitedFromNs, running, startNs, cursorNs, plots, events };
}
