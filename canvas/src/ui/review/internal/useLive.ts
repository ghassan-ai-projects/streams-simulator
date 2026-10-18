import { useCallback, useMemo } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewState, ReviewSession } from "../../../app/review";
import { readyReport } from "../../../app/review";
import type { PlantView } from "../../../model/graph";
import type { NodeLive } from "../../graph";
import { effectorAt, faultedAt, hallLiveOf } from "../../../model/review";

export function useLive(
  review: ReviewSession,
  state: ReviewState,
  plant: ReadyPlant | null,
  view: PlantView | null,
): (id: string) => NodeLive {
  const report = readyReport(state);
  const { positionNs, origins } = state;
  const faulted = useMemo(() => faultedAt(report, positionNs), [report, positionNs]);
  const effectors = useMemo(() => effectorAt(report, positionNs), [report, positionNs]);
  return useCallback(
    (id) =>
      hallLiveOf(
        {
          index: review.index,
          plant: plant?.plant ?? null,
          view,
          faulted,
          effectors,
          origins,
          positionNs,
          sessionLive: state.live,
        },
        id,
      ),
    [review, positionNs, plant, view, faulted, effectors, origins, state.live],
  );
}
