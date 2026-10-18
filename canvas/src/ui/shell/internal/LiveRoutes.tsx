import { LiveSessionHold, ReviewScreen } from "../../review";
import { useStore } from "../../primitives";
import type { Stores } from "./stores";

export function ReviewRoute({ stores, worldId }: { readonly stores: Stores; readonly worldId: string }) {
  const live = useStore(stores.liveReview);
  return (
    <ReviewScreen
      review={live.worldId === worldId ? stores.liveReview : stores.review}
      plants={stores.plants}
      workspace={stores.workspace}
      navigation={stores.navigation}
      launch={stores.launch}
      worldId={worldId}
    />
  );
}

export function KeepLiveSession({ stores }: { readonly stores: Stores }) {
  const connected = useStore(stores.launch.connect);
  if (connected.phase !== "live") return null;
  return <LiveSessionHold review={stores.liveReview} workspace={stores.workspace} run={connected.run} />;
}
