import { ReactFlowProvider } from "@xyflow/react";
import { Network } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { hallCount, hallOf, hallsOf, plantView } from "../../model/plantView";
import type { RunRow } from "../../model/workspace";
import { runRows } from "../../model/workspace";
import type { Crumb } from "../shell/Breadcrumb";
import { Breadcrumb } from "../shell/Breadcrumb";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import { Navigator, ZoomBar } from "./GraphControls";
import { LaunchPanel } from "../launch/LaunchPanel";
import { EditButton, RunButton, useRunningHere } from "./PlantActions";
import { Inspector } from "./Inspector";
import { Outline } from "./Outline";
import { useCanvasAspect } from "./canvasAspect";
import { PlantGraph } from "./PlantGraph";
import styles from "./PlantScreen.module.css";

type Ready = Extract<PlantState, { phase: "ready" }>;

interface PlantScreenProps {
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly domain: string;
  readonly hall: string | null;
  readonly startLaunching?: boolean;
  readonly onShowDetails?: () => void;
}

function summary(state: Ready, hall: string | null): string {
  const ports = state.plant.links.some((link) => link.fromPort === null) ? " · some links record no ports" : "";
  const figures =
    hall === null
      ? `${String(state.plant.instances.length)} instances · ${String(state.plant.links.length)} links`
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

function Graph({
  state,
  hall,
  plants,
  navigation,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly plants: PlantStore;
  readonly navigation: NavigationStore;
}) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  const view = useMemo(
    () => plantView(state.plant, state.specs, hall, 0, aspect),
    [state.plant, state.specs, hall, aspect],
  );
  return (
    <div ref={ref} className={styles.graphHost}>
      <PlantGraph
        view={view}
        viewKey={hall ?? "plant"}
        fitKey={`${hall ?? "plant"}:${String(aspect)}`}
        selected={state.selected}
        onSelect={(id) => {
          plants.select(id);
        }}
        onOpenHall={(name) => {
          navigation.go({ screen: "plant", domain: state.path, hall: name });
        }}
        onEscape={() => {
          if (hall !== null) navigation.go({ screen: "plant", domain: state.path, hall: null });
        }}
      />
    </div>
  );
}

function PlantSide({
  state,
  hall,
  plants,
  navigation,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly plants: PlantStore;
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
  readonly navigation: NavigationStore;
  readonly launching: boolean;
  readonly onClose: () => void;
  readonly plants: PlantStore;
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

const noop = () => undefined;

function useConsume(navigation: NavigationStore, domain: string, hall: string | null) {
  return useCallback(() => {
    navigation.replace({ screen: "plant", domain, hall });
  }, [navigation, domain, hall]);
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

function NotReady({ state, domain }: { readonly state: PlantState; readonly domain: string }) {
  const failed = state.phase === "failed" && state.path === domain;
  return (
    <p className={failed ? styles.failure : styles.loading} role={failed ? "alert" : "status"}>
      {state.phase === "failed" && failed
        ? `The plant could not be opened: ${state.failure.message} (${state.failure.code})`
        : "Opening the plant…"}
    </p>
  );
}

export function PlantScreen(props: PlantScreenProps) {
  const { plants, workspace, navigation, launch, domain, hall, startLaunching = false, onShowDetails = noop } = props;
  const run = useLaunching(startLaunching, onShowDetails, useConsume(navigation, domain, hall));
  const state = useStore(plants);
  const running = useRunningHere(launch, state.phase === "ready" ? state.plant.id : "");
  const listing = useStore(workspace);
  useEffect(() => {
    if (listing.phase === "ready") void plants.load(domain, listing.workspace);
  }, [plants, domain, listing]);
  if (state.phase !== "ready" || state.path !== domain) return <NotReady state={state} domain={domain} />;
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
          <Graph state={state} hall={hall} plants={plants} navigation={navigation} />
        </section>
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
