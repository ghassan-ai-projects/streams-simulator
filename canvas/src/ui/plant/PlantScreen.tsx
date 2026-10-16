import { ReactFlowProvider } from "@xyflow/react";
import { Network } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import type { LaunchStore } from "../../app/launchStore";
import type { Scheduler } from "../../app/clock";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantSession } from "../../app/routeModules";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { hallCount, hallOf, hallsOf, plantView } from "../../model/graph";
import type { RunRow } from "../../model/workspace";
import { counted, runRows } from "../../model/workspace";
import type { Crumb } from "../shell/Breadcrumb";
import { Breadcrumb } from "../shell/Breadcrumb";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import { Navigator, Outline, ZoomBar } from "../graph";
import { LaunchPanel } from "../launch/LaunchPanel";
import { EditButton, RunButton, useRunningHere } from "./PlantActions";
import { Inspector } from "./Inspector";
import { PlantFailure } from "./PlantFailure";
import { PlantCanvas } from "./PlantCanvas";
import styles from "./PlantScreen.module.css";

type Ready = Extract<PlantState, { phase: "ready" }>;

interface PlantScreenProps {
  readonly plants: PlantSession;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly scheduler: Scheduler;
  readonly domain: string;
  readonly hall: string | null;
  readonly startLaunching?: boolean;
  readonly onShowDetails?: () => void;
}

function summary(state: Ready, hall: string | null): string {
  const ports = state.plant.links.some((link) => link.fromPort === null) ? " · some links record no ports" : "";
  const figures =
    hall === null
      ? `${counted(state.plant.instances.length, "instance", "instances")} · ${counted(state.plant.links.length, "link", "links")}`
      : hallCount(plantView(state.plant, state.specs, hall), hall);
  return `${figures}${ports}`;
}

function Crumbs({
  state,
  hall,
  navigation,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly navigation: NavigationStore;
}) {
  const title = state.plant.title || state.plant.id;
  const heading = useScreen(hall === null ? `${title} · plant` : `${hall} · ${title}`);
  const trail: Crumb[] = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "home" });
      },
    },
  ];
  if (hall !== null) {
    trail.push({
      label: title,
      onGo: () => {
        navigation.go({ screen: "plant", domain: state.path, hall: null });
      },
    });
  }
  return <Breadcrumb trail={trail} here={hall ?? title} heading={heading} />;
}

function PlantSide({
  state,
  hall,
  plants,
  navigation,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
}) {
  return (
    <div className={styles.side}>
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
  readonly state: Ready;
  readonly hall: string | null;
  readonly launch: LaunchStore;
  readonly scheduler: Scheduler;
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
  const { state, launch, scheduler, navigation, launching, onClose } = props;
  if (!launching) return <PlantInspector {...props} />;
  return (
    <LaunchPanel
      launch={launch}
      scheduler={scheduler}
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

const noop = () => undefined;

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
    <p className={styles.loading} role="status">
      Opening the plant…
    </p>
  );
}

export function PlantScreen(props: PlantScreenProps) {
  const {
    plants,
    workspace,
    navigation,
    launch,
    scheduler,
    domain,
    hall,
    startLaunching = false,
    onShowDetails = noop,
  } = props;
  const run = useLaunching(startLaunching, useShowRun(launch, onShowDetails), useConsume(navigation, domain, hall));
  const state = useStore(plants);
  const running = useRunningHere(launch, state.phase === "ready" ? state.plant.id : "");
  const listing = useStore(workspace);
  useEffect(() => {
    if (listing.phase === "ready") void plants.load(domain, listing.workspace);
  }, [plants, domain, listing]);
  useEffect(() => () => plants.close(), [plants, domain]);
  if (state.phase !== "ready" || state.path !== domain)
    return <NotReady state={state} domain={domain} navigation={navigation} />;
  return (
    <ReactFlowProvider>
      <div className={styles.screen}>
        <PlantSide state={state} hall={hall} plants={plants} navigation={navigation} />
        <section className={styles.center} aria-label={`Plant ${state.plant.id}`}>
          <div className={styles.bar}>
            <Network size={14} aria-hidden="true" />
            <Crumbs state={state} hall={hall} navigation={navigation} />
            <span className={styles.summary}>{summary(state, hall)}</span>
            <EditButton composition={state.composition} running={running} navigation={navigation} />
            <RunButton launching={run.launching} running={running} button={run.button} onToggle={run.toggle} />
            <ZoomBar />
          </div>
          <PlantCanvas state={state} hall={hall} plants={plants} navigation={navigation} launch={launch} />
        </section>
        <RightPanel
          state={state}
          hall={hall}
          launch={launch}
          scheduler={scheduler}
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
