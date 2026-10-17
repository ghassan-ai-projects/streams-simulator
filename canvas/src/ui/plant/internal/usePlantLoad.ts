import { useEffect } from "react";
import type { PlantSession } from "../../../app/plant";
import type { WorkspaceStore } from "../../../app/workspace";
import { useStore } from "../../primitives";

export function usePlantLoad(plants: PlantSession, domain: string, workspace: WorkspaceStore) {
  const listing = useStore(workspace);
  useEffect(() => {
    if (listing.phase === "ready") void plants.load(domain, listing.workspace);
  }, [plants, domain, listing]);
  useEffect(
    () => () => {
      plants.close();
    },
    [plants, domain],
  );
  return listing;
}
