import { useMemo } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewState } from "../../../app/review";
import { shownOrigins } from "../../../model/review";

export function useShownOrigins(state: ReviewState, plant: ReadyPlant | null): ReviewState {
  const { origins, connection, simulated } = state;
  const shown = useMemo(
    () => shownOrigins(origins, plant?.plant.instances.map((instance) => instance.id) ?? [], connection, simulated),
    [origins, plant, connection, simulated],
  );
  return shown === origins ? state : { ...state, origins: shown };
}
