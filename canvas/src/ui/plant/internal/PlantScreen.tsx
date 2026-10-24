import type { ReactNode } from "react";
import { screen } from "../../primitives";
import { ReactFlowProvider } from "@xyflow/react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { PlantMaker, PlantState, PlantSession, ReadyPlant } from "../../../app/plant";
import type { ProvenanceSession, WorkspaceStore } from "../../../app/workspace";
import { hallOf, hallsOf } from "../../../model/graph";
import { plantTitle } from "../../../model/plant";
import type { RunRow } from "../../../model/workspace";
import type { Workspace } from "../../../model/workspace";
import { runRows } from "../../../model/workspace";
import { useStore } from "../../primitives";
import { Navigator, Outline } from "../../graph";
import { provenanceParts, useDrawer } from "./provenanceParts";
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
  readonly provenance: ProvenanceSession;
  readonly maker: PlantMaker;
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
  block,
}: {
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
  readonly connected?: Connected;
  readonly block: ReactNode;
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
      {block}
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
  readonly provenance: ReactNode;
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
  const { state, launch, navigation, launching, onClose, provenance } = props;
  if (!launching && provenance !== null) return provenance;
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

interface ReadyProps {
  readonly props: PlantScreenProps;
  readonly state: ReadyPlant;
  readonly workspace: Workspace | null;
  readonly run: ReturnType<typeof useLaunching>;
  readonly drawer: ReturnType<typeof useDrawer>;
}

function Ready({ props, state, workspace, run, drawer }: ReadyProps) {
  const { plants, navigation, launch, hall } = props;
  const childProps = {
    state,
    hall,
    plants,
    navigation,
    ...(props.connected === undefined ? {} : { connected: props.connected }),
  };
  const parts = provenanceParts({
    provenance: props.provenance,
    navigation,
    maker: props.maker,
    path: state.path,
    title: plantTitle(state.plant),
    workspace,
    drawer,
  });
  return (
    <ReactFlowProvider>
      <div className={screen.screen}>
        <PlantSide {...childProps} block={parts.block} />
        <PlantCentre {...childProps} launch={launch} run={run} />
        <RightPanel
          {...childProps}
          launch={launch}
          launching={run.launching}
          provenance={parts.panel}
          onClose={run.close}
          runs={workspace === null ? [] : runRows(workspace, state.plant.id)}
        />
      </div>
    </ReactFlowProvider>
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
  const drawer = useDrawer(onShowDetails);
  const listing = usePlantLoad(plants, domain, workspace);
  if (state.phase !== "ready" || state.path !== domain)
    return <NotReady state={state} domain={domain} navigation={navigation} />;
  const opened = listing.phase === "ready" ? listing.workspace : null;
  return <Ready props={props} state={state} workspace={opened} run={run} drawer={drawer} />;
}
