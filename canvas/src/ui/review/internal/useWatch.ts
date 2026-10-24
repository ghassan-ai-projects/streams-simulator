import { useMemo } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession } from "../../../app/review";
import { isWatched, watchLimit } from "../../../model/review";
import { useStore } from "../../primitives";
import type { WatchControls } from "./WatchButton";

const nothing: readonly never[] = [];

export function useWatch(review: ReviewSession, plant: ReadyPlant | null, entity: string): WatchControls | null {
  const state = useStore(review.watchList);
  const plantId = plant?.plant.id;
  const list = state.lists.get(plantId ?? "") ?? nothing;
  return useMemo(() => {
    if (plantId === undefined) return null;
    return {
      watched: (channel: string) => isWatched(list, { entity, channel }),
      full: list.length >= watchLimit,
      toggle: (channel: string) => {
        review.watchList.toggle(plantId, { entity, channel });
      },
    };
  }, [review, plantId, list, entity]);
}
