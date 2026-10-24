import { useCallback, useMemo } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewState, ReviewSession } from "../../../app/review";
import { readyReport, runGoing } from "../../../app/review";
import type { PlantView } from "../../../model/graph";
import type { NodeLive } from "../../graph";
import { effectorAt, faultedAt, hallLiveOf } from "../../../model/review";
import { useShownLive } from "./useShownLive";

export function useLive(
  review: ReviewSession,
  state: ReviewState,
  plant: ReadyPlant | null,
  view: PlantView | null,
): (id: string) => NodeLive {
  const report = readyReport(state);
  const { positionNs, origins } = state;
  const sessionLive = useShownLive(review, state);
  const phase = runGoing(state) ? "live" : "ended";
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
          sessionLive,
          phase,
        },
        id,
      ),
    [review, positionNs, plant, view, faulted, effectors, origins, sessionLive, phase],
  );
}
