import { ReactFlowProvider } from "@xyflow/react";
import { ChevronRight } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { stateUnit } from "../../model/plant";
import { plantView } from "../../model/plantView";
import type { Report } from "../../model/report";
import { runSpan } from "../../model/report";
import { Navigator, ZoomBar } from "../plant/GraphControls";
import { Outline } from "../plant/Outline";
import { readoutHeight } from "../plant/nodes";
import { PlantGraph } from "../plant/PlantGraph";
import shared from "../shared.module.css";
import { useScreen } from "../useScreen";
import { useStore } from "../useStore";
import { Banners, Status } from "./RunStatus";
import { RunPanel } from "./RunPanel";
import { Scrubber } from "./Scrubber";
import { useLive } from "./useLive";
import styles from "./RunScreen.module.css";

interface RunScreenProps {
  readonly review: ReviewStore;
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly worldId: string;
}

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

interface HallProps {
  readonly hall: string | null;
  readonly onHall: (hall: string | null) => void;
}

function RunCrumb({ worldId, onHall }: { readonly worldId: string; readonly onHall: (hall: null) => void }) {
  return (
    <>
      <button
        type="button"
        className={shared.crumb}
        onClick={() => {
          onHall(null);
        }}
      >
        {worldId}
      </button>
      <ChevronRight size={14} aria-hidden="true" />
    </>
  );
}

function Header({
  state,
  navigation,
  worldId,
  hall,
  onHall,
}: {
  readonly state: ReviewState;
  readonly navigation: NavigationStore;
  readonly worldId: string;
} & HallProps) {
  const heading = useScreen(hall === null ? `${worldId} · review` : `${hall} · ${worldId} · review`);
  return (
    <div className={styles.bar}>
      <button
        type="button"
        className={shared.crumb}
        onClick={() => {
          navigation.go({ screen: "home" });
        }}
      >
        Workspace
      </button>
      <ChevronRight size={14} aria-hidden="true" />
      {hall !== null && <RunCrumb worldId={worldId} onHall={onHall} />}
      <h1 className={styles.here} ref={heading} tabIndex={-1}>
        {hall ?? worldId}
      </h1>
      <Status state={state} />
    </div>
  );
}

function useRunPlant(state: ReviewState, plants: PlantStore, workspace: WorkspaceStore): ReadyPlant | null {
  const listing = useStore(workspace);
  const plant = useStore(plants);
  const key = `run:${state.worldId ?? ""}`;
  const domainId = state.report.phase === "ready" ? state.report.value.domainId : null;
  const fallback =
    listing.phase === "ready" ? listing.workspace.domains.find((entry) => entry.id === domainId)?.path : undefined;
  useEffect(() => {
    if (state.runDomain.phase === "ready") void plants.show(key, state.runDomain.value);
    else if (state.runDomain.phase === "failed" && fallback !== undefined && listing.phase === "ready")
      void plants.load(fallback, listing.workspace);
  }, [state.runDomain, key, fallback, listing, plants]);
  if (plant.phase !== "ready") return null;
  return plant.path === key || plant.path === fallback ? plant : null;
}

function OpeningPlant() {
  return (
    <p className={styles.muted} role="status">
      Opening the plant this run used…
    </p>
  );
}

function Graph({
  review,
  state,
  plant,
  hall,
  onHall,
}: {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
} & HallProps) {
  const view = useMemo(
    () => (plant === null ? null : plantView(plant.plant, plant.specs, hall, readoutHeight)),
    [plant, hall],
  );
  const live = useLive(review, state, plant);
  if (view === null) return <OpeningPlant />;
  return (
    <div className={styles.graphArea}>
      <div className={styles.graphBar}>
        <span className={styles.muted}>Plant at the playback time · values are observations</span>
        <ZoomBar />
      </div>
      <PlantGraph
        minFitZoom={0.5}
        view={view}
        viewKey={hall ?? "review"}
        selected={state.selected}
        live={live}
        onSelect={(id) => {
          review.select(id);
        }}
        onOpenHall={onHall}
        onEscape={() => {
          onHall(null);
        }}
      />
    </div>
  );
}

function Timeline({ review, state }: { readonly review: ReviewStore; readonly state: ReviewState }) {
  const report: Report | null = state.report.phase === "ready" ? state.report.value : null;
  const span = runSpan(report, review.index.bounds());
  return (
    <Scrubber
      startNs={span.startNs}
      endNs={span.endNs}
      positionNs={state.positionNs === 0 ? span.endNs : state.positionNs}
      report={report}
      onSeek={(ns) => {
        review.seek(ns);
      }}
      onEvent={(mark) => {
        review.reveal(mark.atNs, mark.entity);
      }}
    />
  );
}

function Side({
  review,
  state,
  plant,
  hall,
  onHall,
}: {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
} & HallProps) {
  return (
    <div className={styles.side}>
      {plant !== null && (
        <Outline
          plant={plant.plant}
          hall={hall}
          selected={state.selected}
          onSelect={(id) => {
            review.select(id);
          }}
          onOpenHall={onHall}
        />
      )}
      <Navigator />
    </div>
  );
}

export function RunScreen({ review, plants, workspace, navigation, worldId }: RunScreenProps) {
  const state = useStore(review);
  const plant = useRunPlant(state, plants, workspace);
  const [hall, setHall] = useState<string | null>(null);
  useEffect(() => {
    review.open(worldId);
  }, [review, worldId]);
  const unit = useCallback(
    (entity: string, name: string) => (plant === null ? "" : stateUnit(plant.plant, entity, name)),
    [plant],
  );
  return (
    <ReactFlowProvider>
      <div className={styles.screen}>
        <Header state={state} navigation={navigation} worldId={worldId} hall={hall} onHall={setHall} />
        <Side review={review} state={state} plant={plant} hall={hall} onHall={setHall} />
        <section className={styles.center} aria-label={`Run ${worldId}`}>
          <Banners state={state} />
          <Timeline review={review} state={state} />
          <Graph review={review} state={state} plant={plant} hall={hall} onHall={setHall} />
        </section>
        <RunPanel review={review} state={state} plant={plant} unit={unit} />
      </div>
    </ReactFlowProvider>
  );
}
