import { ReactFlowProvider } from "@xyflow/react";
import { ChevronRight, Network } from "lucide-react";
import { Fragment, useEffect, useMemo } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { plantView } from "../../model/plantView";
import type { Route } from "../../model/route";
import { useStore } from "../useStore";
import { Navigator, ZoomBar } from "./GraphControls";
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
  readonly domain: string;
  readonly hall: string | null;
}

interface Crumb {
  readonly label: string;
  readonly route: Route;
}

function summary(state: Ready): string {
  const ports = state.composition === null ? " · no composition, ports unknown" : "";
  return `${String(state.plant.instances.length)} instances · ${String(state.plant.links.length)} links${ports}`;
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
      <h1 aria-current="location" className={styles.here}>
        {hall ?? title}
      </h1>
      <span className={styles.summary}>{summary(state)}</span>
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
    />
  );
}

export function PlantScreen({ plants, workspace, navigation, domain, hall }: PlantScreenProps) {
  const state = useStore(plants);
  const listing = useStore(workspace);
  useEffect(() => {
    if (listing.phase === "ready") void plants.load(domain, listing.workspace);
  }, [plants, domain, listing]);
  if (state.phase !== "ready" || state.path !== domain) {
    const failed = state.phase === "failed" && state.path === domain;
    return (
      <p className={failed ? styles.failure : styles.loading} role={failed ? "alert" : "status"}>
        {failed
          ? `The plant could not be opened: ${state.failure.message} (${state.failure.code})`
          : "Opening the plant…"}
      </p>
    );
  }
  return (
    <ReactFlowProvider>
      <div className={styles.screen}>
        <div className={styles.side}>
          <Outline state={state} hall={hall} plants={plants} navigation={navigation} />
          <Navigator />
        </div>
        <section className={styles.center} aria-label={`Plant ${state.plant.id}`}>
          <div className={styles.bar}>
            <Network size={14} aria-hidden="true" />
            <Crumbs state={state} hall={hall} navigation={navigation} />
            <ZoomBar />
          </div>
          <Graph state={state} hall={hall} plants={plants} navigation={navigation} />
        </section>
        <Inspector plant={state.plant} selected={state.selected} />
      </div>
    </ReactFlowProvider>
  );
}
