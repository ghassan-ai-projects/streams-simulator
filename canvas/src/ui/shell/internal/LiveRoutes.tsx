import type { WorkspaceState } from "../../../app/workspace";
import { LiveSessionHold, ReviewScreen } from "../../review";
import { useStore } from "../../primitives";
import type { Stores } from "./stores";

function listedRunning(listing: WorkspaceState, worldId: string): boolean {
  if (listing.phase !== "ready") return false;
  return listing.workspace.runs.some((run) => run.worldId === worldId && run.connected === "running");
}

export function ReviewRoute({
  stores,
  worldId,
  model,
}: {
  readonly stores: Stores;
  readonly worldId: string;
  readonly model: boolean;
}) {
  const held = useStore(stores.liveReview);
  const connected = useStore(stores.launch.connect);
  const listing = useStore(stores.workspace);
  const live = (connected.phase === "live" && connected.run === worldId) || listedRunning(listing, worldId);
  return (
    <ReviewScreen
      live={live}
      review={live || held.worldId === worldId ? stores.liveReview : stores.review}
      plants={stores.plants}
      workspace={stores.workspace}
      navigation={stores.navigation}
      launch={stores.launch}
      provenance={stores.provenance}
      model={model}
      worldId={worldId}
    />
  );
}

export function KeepLiveSession({ stores }: { readonly stores: Stores }) {
  const connected = useStore(stores.launch.connect);
  if (connected.phase !== "live") return null;
  return <LiveSessionHold review={stores.liveReview} workspace={stores.workspace} run={connected.run} />;
}
