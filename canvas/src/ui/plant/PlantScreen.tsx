import { ReactFlowProvider } from "@xyflow/react";
import { ChevronRight, Network, PenLine, Play } from "lucide-react";
import { Fragment, useEffect, useMemo, useState } from "react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { hallCount, plantView } from "../../model/plantView";
import type { Route } from "../../model/route";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import { Navigator, ZoomBar } from "./GraphControls";
import { LaunchPanel } from "../launch/LaunchPanel";
import { Inspector } from "./Inspector";
import { Outline } from "./Outline";
import { PlantGraph } from "./PlantGraph";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

type Ready = Extract<PlantState, { phase: "ready" }>;

interface PlantScreenProps {
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly domain: string;
  readonly hall: string | null;
}

interface Crumb {
  readonly label: string;
  readonly route: Route;
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
  const trail: readonly Crumb[] = [
    { label: "Workspace", route: { screen: "home" } },
    ...(hall === null ? [] : [{ label: title, route: { screen: "plant", domain: state.path, hall: null } } as const]),
  ];
  return (
    <nav className={styles.crumbs} aria-label="Breadcrumb">
      {trail.map((crumb) => (
        <Fragment key={crumb.label}>
          <button
            type="button"
            className={shared.crumb}
            onClick={() => {
              navigation.go(crumb.route);
            }}
          >
            {crumb.label}
          </button>
          <ChevronRight size={14} aria-hidden="true" />
        </Fragment>
      ))}
      <h1 aria-current="location" className={styles.here} ref={heading} tabIndex={-1}>
        {hall ?? title}
      </h1>
      <span className={styles.summary}>{summary(state, hall)}</span>
    </nav>
  );
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
  const view = useMemo(() => plantView(state.plant, state.specs, hall), [state.plant, state.specs, hall]);
  return (
    <PlantGraph
      view={view}
      viewKey={hall ?? "plant"}
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

function RightPanel({
  state,
  launch,
  navigation,
  launching,
  setLaunching,
}: {
  readonly state: Ready;
  readonly launch: LaunchStore;
  readonly navigation: NavigationStore;
  readonly launching: boolean;
  readonly setLaunching: (open: boolean) => void;
}) {
  if (!launching) return <Inspector plant={state.plant} selected={state.selected} />;
  return (
    <LaunchPanel
      launch={launch}
      plant={state.plant}
      onClose={() => {
        setLaunching(false);
      }}
      onFinished={(worldId) => {
        navigation.go({ screen: "run", worldId });
      }}
    />
  );
}

function RunButton({ launching, onOpen }: { readonly launching: boolean; readonly onOpen: () => void }) {
  return (
    <button type="button" className={styles.run} aria-pressed={launching} onClick={onOpen}>
      <Play size={14} aria-hidden="true" />
      Run plant
    </button>
  );
}

function EditButton({ state, navigation }: { readonly state: Ready; readonly navigation: NavigationStore }) {
  const composition = state.composition;
  if (composition === null) return null;
  return (
    <button
      type="button"
      className={styles.edit}
      onClick={() => {
        navigation.go({ screen: "edit", composition });
      }}
    >
      <PenLine size={14} aria-hidden="true" />
      Edit plant
    </button>
  );
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

export function PlantScreen({ plants, workspace, navigation, launch, domain, hall }: PlantScreenProps) {
  const [launching, setLaunching] = useState(false);
  const state = useStore(plants);
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
            <EditButton state={state} navigation={navigation} />
            <RunButton
              launching={launching}
              onOpen={() => {
                setLaunching(true);
              }}
            />
            <ZoomBar />
          </div>
          <Graph state={state} hall={hall} plants={plants} navigation={navigation} />
        </section>
        <RightPanel
          state={state}
          launch={launch}
          navigation={navigation}
          launching={launching}
          setLaunching={setLaunching}
        />
      </div>
    </ReactFlowProvider>
  );
}
