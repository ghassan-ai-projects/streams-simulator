import type { Route } from "../../../model/shell";
import { PlantScreen } from "../../plant";
import { ConnectedPlant } from "../../review";
import { useStore } from "../../primitives";
import type { Stores } from "./stores";

export function PlantRoute({
  stores,
  route,
}: {
  readonly stores: Stores;
  readonly route: Extract<Route, { screen: "plant" }>;
}) {
  const plant = useStore(stores.plants);
  const connected = useStore(stores.launch.connect);
  const ready = plant.phase === "ready" && plant.path === route.domain ? plant : null;
  const run = connected.phase === "live" && connected.domain === ready?.plant.id ? connected.run : null;
  const props = {
    plants: stores.plants,
    workspace: stores.workspace,
    navigation: stores.navigation,
    launch: stores.launch,
    domain: route.domain,
    hall: route.hall,
    startLaunching: route.launch === true,
    onShowDetails: stores.showDetails,
  };
  return (
    <ConnectedPlant
      review={stores.liveReview}
      plant={ready}
      run={run}
      connect={stores.launch.connect}
      now={() => stores.launch.nowMs()}
    >
      {(session) => <PlantScreen {...props} {...(session === null ? {} : { connected: session })} />}
    </ConnectedPlant>
  );
}
