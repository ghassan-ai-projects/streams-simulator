import type { Route } from "../../../model/shell";
import { PlantScreen } from "../../plant";
import { ConnectedPlant } from "../../review";
import { useStore } from "../../primitives";
import type { Stores } from "./App";

export function PlantRoute({
  stores,
  route,
}: {
  readonly stores: Stores;
  readonly route: Extract<Route, { screen: "plant" }>;
}) {
  const plant = useStore(stores.plants);
  const connected = useStore(stores.launch.connect);
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
  if (connected.phase !== "live") return <PlantScreen {...props} />;
  const own = plant.phase === "ready" && plant.path === route.domain && connected.domain === plant.plant.id;
  return (
    <ConnectedPlant
      review={stores.review}
      workspace={stores.workspace}
      plant={plant.phase === "ready" ? plant : null}
      run={connected.run}
    >
      {(session) => <PlantScreen {...props} {...(own ? { connected: session } : {})} />}
    </ConnectedPlant>
  );
}
