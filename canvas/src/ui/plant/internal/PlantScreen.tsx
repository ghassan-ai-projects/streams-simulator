import { screen } from "../../primitives";
import { ReactFlowProvider } from "@xyflow/react";
import { useCallback, useEffect, useRef, useState } from "react";
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

interface PlantScreenProps {
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
}: {
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
}) {
  return (
    <div className={screen.side}>
      <Outline
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
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly launch: LaunchStore;
  readonly navigation: NavigationStore;
  readonly launching: boolean;
  readonly onClose: () => void;
  readonly plants: PlantSession;
  readonly runs: readonly RunRow[];
}

function PlantInspector({ state, hall, navigation, plants, runs }: RightProps) {
  return (
    <Inspector
      plant={state.plant}
      specs={state.specs}
      selected={state.selected}
      runs={runs}
      onClose={() => {
        plants.select(null);
      }}
      onReview={(worldId) => {
        navigation.go({ screen: "run", worldId });
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
        navigation.go({ screen: "run", worldId });
      }}
    />
  );
}

function useConsume(navigation: NavigationStore, domain: string, hall: string | null) {
  return useCallback(() => {
    navigation.replace({ screen: "plant", domain, hall });
  }, [navigation, domain, hall]);
}

function useShowRun(launch: LaunchStore, onShow: () => void) {
  return useCallback(() => {
    launch.reset();
    onShow();
  }, [launch, onShow]);
}

function useLaunching(asked: boolean, onShow: () => void, consume: () => void) {
  const [launching, setLaunching] = useState(asked);
  const button = useRef<HTMLButtonElement>(null);
  const [seen, setSeen] = useState(asked);
  if (asked !== seen) {
    setSeen(asked);
    if (asked) setLaunching(true);
  }
  useEffect(() => {
    if (!asked) return;
    onShow();
    consume();
  }, [asked, onShow, consume]);
  const toggle = () => {
    if (!launching) onShow();
    setLaunching(!launching);
  };
  const close = () => {
    setLaunching(false);
    button.current?.focus();
  };
  return { launching, button, toggle, close };
}

function usePlantLoad(plants: PlantSession, domain: string, workspace: WorkspaceStore) {
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
  const run = useLaunching(startLaunching, useShowRun(launch, onShowDetails), useConsume(navigation, domain, hall));
  const state = useStore(plants);
  const listing = usePlantLoad(plants, domain, workspace);
  if (state.phase !== "ready" || state.path !== domain)
    return <NotReady state={state} domain={domain} navigation={navigation} />;
  return (
    <ReactFlowProvider>
      <div className={screen.screen}>
        <PlantSide state={state} hall={hall} plants={plants} navigation={navigation} />
        <PlantCentre state={state} hall={hall} plants={plants} navigation={navigation} launch={launch} run={run} />
        <RightPanel
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
