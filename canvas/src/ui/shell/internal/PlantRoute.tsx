import { watchedWorld } from "../../../app/launch";
import type { ReadyPlant } from "../../../app/plant";
import type { Route } from "../../../model/shell";
import { PlantScreen } from "../../plant";
import { ConnectedPlant, RunWatchHold } from "../../review";
import { useStore } from "../../primitives";
import type { Stores } from "./stores";

interface Following {
  readonly kind: "session" | "watch";
  readonly run: string | null;
}

function useFollowing(stores: Stores, ready: ReadyPlant | null): Following {
  const connected = useStore(stores.launch.connect);
  const launched = useStore(stores.launch);
  const session = connected.phase === "live" && connected.domain === ready?.plant.id ? connected.run : null;
  if (session !== null || ready === null) return { kind: "session", run: session };
  return { kind: "watch", run: watchedWorld(launched, ready.plant.id) };
}

export function PlantRoute({
  stores,
  route,
}: {
  readonly stores: Stores;
  readonly route: Extract<Route, { screen: "plant" }>;
}) {
  const plant = useStore(stores.plants);
  const ready = plant.phase === "ready" && plant.path === route.domain ? plant : null;
  const { kind, run } = useFollowing(stores, ready);
  const props = {
    plants: stores.plants,
    workspace: stores.workspace,
    navigation: stores.navigation,
    launch: stores.launch,
    provenance: stores.provenance,
    maker: stores.maker,
    domain: route.domain,
    hall: route.hall,
    startLaunching: route.launch === true,
    onShowDetails: stores.showDetails,
  };
  return (
    <>
      {kind === "watch" && run !== null && (
        <RunWatchHold review={stores.watchReview} workspace={stores.workspace} run={run} />
      )}
      <ConnectedPlant
        kind={kind}
        review={kind === "watch" ? stores.watchReview : stores.liveReview}
        plant={ready}
        run={run}
        connect={stores.launch.connect}
        now={() => stores.launch.nowMs()}
      >
        {(live) => <PlantScreen {...props} {...(live === null ? {} : { connected: live })} />}
      </ConnectedPlant>
    </>
  );
}
