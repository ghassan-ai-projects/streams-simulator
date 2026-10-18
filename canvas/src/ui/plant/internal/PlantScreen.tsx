import { screen } from "../../primitives";
import { ReactFlowProvider } from "@xyflow/react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { PlantState, PlantSession, ReadyPlant } from "../../../app/plant";
import type { WorkspaceStore } from "../../../app/workspace";
import { hallOf, hallsOf } from "../../../model/graph";
import type { RunRow } from "../../../model/workspace";
import { runRows } from "../../../model/workspace";
import { useStore } from "../../primitives";
import { Navigator, Outline } from "../../graph";
import { LaunchPanel } from "../../launch";
import { Inspector } from "./Inspector";
import { PlantCentre } from "./PlantCentre";
import { PlantFailure } from "./PlantFailure";
import { useConsume, useLaunching, useShowRun } from "./useLaunching";
import { usePlantLoad } from "./usePlantLoad";
import type { Connected } from "./connected";
import { ConnectedReadings } from "./ConnectedReadings";

interface PlantScreenProps {
  readonly connected?: Connected;
  readonly plants: PlantSession;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly domain: string;
  readonly hall: string | null;
  readonly startLaunching: boolean;
  readonly onShowDetails: () => void;
}

function PlantSide({
  state,
  hall,
  plants,
  navigation,
  connected,
}: {
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
  readonly connected?: Connected;
}) {
  return (
    <div className={screen.side}>
      <Outline
        {...(connected === undefined ? {} : { origins: connected.origins })}
        plant={state.plant}
        hall={hall}
        selected={state.selected}
        onSelect={(id) => {
          plants.select(id);
        }}
        onOpenHall={(name) => {
          navigation.go({ screen: "plant", domain: state.path, hall: name });
        }}
      />
      <Navigator />
    </div>
  );
}

interface RightProps {
  readonly connected?: Connected;
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly launch: LaunchStore;
  readonly navigation: NavigationStore;
  readonly launching: boolean;
  readonly onClose: () => void;
  readonly plants: PlantSession;
  readonly runs: readonly RunRow[];
}

function PlantInspector({ state, hall, navigation, plants, runs, connected }: RightProps) {
  return (
    <Inspector
      liveDetail={
        connected !== undefined && state.selected !== null ? (
          <ConnectedReadings connected={connected} entity={state.selected} />
        ) : null
      }
      plant={state.plant}
      specs={state.specs}
      selected={state.selected}
      runs={runs}
      onClose={() => {
        plants.select(null);
      }}
      onReview={(worldId) => {
        navigation.go({ screen: "review", worldId });
      }}
      onSelect={(id) => {
        if (!state.plant.instances.some((instance) => instance.id === id)) return;
        plants.select(id);
        const destination = hallOf(id);
        if (destination !== null && hallsOf(state.plant).includes(destination) && destination !== hall) {
          navigation.go({ screen: "plant", domain: state.path, hall: destination });
        }
      }}
    />
  );
}

function RightPanel(props: RightProps) {
  const { state, launch, navigation, launching, onClose } = props;
  if (!launching) return <PlantInspector {...props} />;
  return (
    <LaunchPanel
      launch={launch}
      plant={state.plant}
      path={state.path}
      selected={state.selected}
      onClose={onClose}
      onFinished={(worldId) => {
        navigation.go({ screen: "review", worldId });
      }}
    />
  );
}

function NotReady({
  state,
  domain,
  navigation,
}: {
  readonly state: PlantState;
  readonly domain: string;
  readonly navigation: NavigationStore;
}) {
  if (state.phase === "failed" && state.path === domain)
    return <PlantFailure failure={state.failure} navigation={navigation} />;
  return (
    <p className={screen.loading} role="status">
      Opening the plant…
    </p>
  );
}

export function PlantScreen(props: PlantScreenProps) {
  const { plants, workspace, navigation, launch, domain, hall, startLaunching, onShowDetails } = props;
  const run = useLaunching(
    { asked: startLaunching },
    useShowRun(launch, onShowDetails),
    useConsume(navigation, domain, hall),
  );
  const state = useStore(plants);
  const listing = usePlantLoad(plants, domain, workspace);
  if (state.phase !== "ready" || state.path !== domain)
    return <NotReady state={state} domain={domain} navigation={navigation} />;
  const childProps = {
    state,
    hall,
    plants,
    navigation,
    ...(props.connected === undefined ? {} : { connected: props.connected }),
  };
  return (
    <ReactFlowProvider>
      <div className={screen.screen}>
        <PlantSide {...childProps} />
        <PlantCentre {...childProps} launch={launch} run={run} />
        <RightPanel
          {...(props.connected === undefined ? {} : { connected: props.connected })}
          state={state}
          hall={hall}
          launch={launch}
          navigation={navigation}
          launching={run.launching}
          onClose={run.close}
          plants={plants}
          runs={listing.phase === "ready" ? runRows(listing.workspace, state.plant.id) : []}
        />
      </div>
    </ReactFlowProvider>
  );
}
