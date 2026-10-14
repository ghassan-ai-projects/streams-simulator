import { ReactFlowProvider } from "@xyflow/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { stateUnit } from "../../model/plant";
import { plantView } from "../../model/plantView";
import type { Report } from "../../model/report";
import { runSpan } from "../../model/report";
import { grouped } from "../../model/units";
import { Navigator, ZoomBar } from "../plant/GraphControls";
import { Outline } from "../plant/Outline";
import { readoutHeight } from "../plant/nodes";
import { PlantGraph } from "../plant/PlantGraph";
import { useStore } from "../useStore";
import { Banners } from "./RunStatus";
import type { HallProps } from "./RunHeader";
import { Header } from "./RunHeader";
import { RunPanel } from "./RunPanel";
import { Scrubber } from "./Scrubber";
import { useLive } from "./useLive";
import { usePlayback } from "./usePlayback";
import styles from "./RunScreen.module.css";

interface RunScreenProps {
  readonly review: ReviewStore;
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly worldId: string;
}

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

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

function graphWords(state: ReviewState, settled: number | null): string {
  const loading = state.playback.phase === "loading" ? " · loading" : "";
  const opening =
    settled !== null && state.positionNs === settled ? " · opens where every instrument has reported" : "";
  return `Plant at the playback time · ${grouped(state.playback.records)} observations${loading}${opening}`;
}

function GraphBar({ state, settled }: { readonly state: ReviewState; readonly settled: number | null }) {
  return (
    <div className={styles.graphBar}>
      <span className={styles.muted}>{graphWords(state, settled)}</span>
      <ZoomBar />
    </div>
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
      <GraphBar state={state} settled={review.index.settled()} />
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
      positionNs={state.positionNs === 0 ? span.startNs : state.positionNs}
      report={report}
      playing={state.playing}
      speed={state.speed}
      onPlay={() => {
        review.play();
      }}
      onPause={() => {
        review.pause();
      }}
      onSpeed={(speed) => {
        review.setSpeed(speed);
      }}
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

function usePlantIdentity(state: ReviewState, workspace: WorkspaceStore, plant: ReadyPlant | null) {
  const listing = useStore(workspace);
  const domainId = state.report.phase === "ready" ? state.report.value.domainId : "";
  const entry = listing.phase === "ready" ? listing.workspace.domains.find((d) => d.id === domainId) : undefined;
  const title = plant?.plant.title ?? entry?.title ?? domainId;
  return { title: title === "" ? "Plant" : title, path: entry?.path ?? null };
}

export function RunScreen({ review, plants, workspace, navigation, launch, worldId }: RunScreenProps) {
  const state = useStore(review);
  const plant = useRunPlant(state, plants, workspace);
  const identity = usePlantIdentity(state, workspace, plant);
  const [hall, setHall] = useState<string | null>(null);
  useEffect(() => {
    review.open(worldId);
    return () => {
      review.close();
    };
  }, [review, worldId]);
  usePlayback(review, state.playing);
  const unit = useCallback(
    (entity: string, name: string) => (plant === null ? "" : stateUnit(plant.plant, entity, name)),
    [plant],
  );
  return (
    <ReactFlowProvider>
      <div className={styles.screen}>
        <Header
          state={state}
          navigation={navigation}
          launch={launch}
          worldId={worldId}
          hall={hall}
          onHall={setHall}
          plant={identity}
        />
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
