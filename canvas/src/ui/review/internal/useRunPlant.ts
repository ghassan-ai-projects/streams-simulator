import { useCallback, useEffect } from "react";
import type { PlantSession, ReadyPlant } from "../../../app/plant";
import type { ReviewPlant, ReviewState } from "../../../app/review";
import { reviewPlantFor, showRunPlant } from "../../../app/review";
import type { WorkspaceStore } from "../../../app/workspace";
import { stateUnit } from "../../../model/plant";
import { useStore } from "../../primitives";

export function useClosePlant(plants: PlantSession, worldId: string): void {
  useEffect(
    () => () => {
      plants.close();
    },
    [plants, worldId],
  );
}

export function useRunPlant(state: ReviewState, plants: PlantSession, workspace: WorkspaceStore): ReviewPlant {
  const listing = useStore(workspace);
  const shown = useStore(plants);
  const known = listing.phase === "ready" ? listing.workspace : null;
  const view = reviewPlantFor(state, shown, known);
  const { runDomain, connection } = state;
  const { key, fallback } = view;
  useEffect(() => {
    showRunPlant(plants, { key, fallback }, { runDomain, connection }, known);
  }, [runDomain, connection, key, fallback, known, plants]);
  return view;
}

export function useUnitOf(plant: ReadyPlant | null) {
  return useCallback(
    (entity: string, name: string) => (plant === null ? "" : stateUnit(plant.plant, entity, name)),
    [plant],
  );
}
