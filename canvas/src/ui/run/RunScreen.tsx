import { ReactFlowProvider } from "@xyflow/react";
import { ChevronRight } from "lucide-react";
import { useCallback, useEffect, useMemo } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { instrumentOf, stateUnit } from "../../model/plant";
import { shortChannel } from "../../model/playback";
import { plantView } from "../../model/plantView";
import { attentionOf, showReading } from "../../model/readout";
import type { Report } from "../../model/report";
import { runSpan } from "../../model/report";
import { faultedAt } from "../../model/timeline";
import type { NodeLive } from "../plant/flow";
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
import styles from "./RunScreen.module.css";

interface RunScreenProps {
  readonly review: ReviewStore;
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly worldId: string;
}

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

function Header({
  state,
  navigation,
  worldId,
}: {
  readonly state: ReviewState;
  readonly navigation: NavigationStore;
  readonly worldId: string;
}) {
  const heading = useScreen(`${worldId} · review`);
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
      <h1 className={styles.here} ref={heading} tabIndex={-1}>
        {worldId}
      </h1>
      <Status state={state} />
      <ZoomBar />
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

function useLive(review: ReviewStore, state: ReviewState, plant: ReadyPlant | null) {
  const report = state.report.phase === "ready" ? state.report.value : null;
  const faulted = useMemo(() => faultedAt(report, state.positionNs), [report, state.positionNs]);
  return useCallback(
    (id: string): NodeLive => {
      const shown = review.index
        .readings(id, state.positionNs)
        .map((reading) =>
          showReading(reading, plant === null ? "" : instrumentOf(plant.plant, id, shortChannel(reading.channel))),
        );
      const attention = attentionOf(shown);
      return { shown, status: attention === "alarm" ? "alarm" : faulted.has(id) ? "fault" : attention };
    },
    [review, state.positionNs, plant, faulted],
  );
}

function Graph({
  review,
  state,
  plant,
}: {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
}) {
  const view = useMemo(
    () => (plant === null ? null : plantView(plant.plant, plant.specs, null, readoutHeight)),
    [plant],
  );
  const live = useLive(review, state, plant);
  if (view === null) {
    return (
      <p className={styles.muted} role="status">
        Opening the plant this run used…
      </p>
    );
  }
  return (
    <PlantGraph
      view={view}
      viewKey="review"
      selected={state.selected}
      live={live}
      onSelect={(id) => {
        review.select(id);
      }}
    />
  );
}

function Timeline({ review, state }: { readonly review: ReviewStore; readonly state: ReviewState }) {
  const report: Report | null = state.report.phase === "ready" ? state.report.value : null;
  const span = runSpan(report, review.index.bounds());
  const density = useMemo(
    () => (state.playback.phase === "ready" ? review.index.density(span.startNs, span.endNs, 120) : []),
    [review, state.playback.phase, span.startNs, span.endNs],
  );
  return (
    <Scrubber
      startNs={span.startNs}
      endNs={span.endNs}
      positionNs={state.positionNs === 0 ? span.endNs : state.positionNs}
      report={report}
      density={density}
      onSeek={(ns) => {
        review.seek(ns);
      }}
    />
  );
}

function Side({
  review,
  state,
  plant,
}: {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
}) {
  return (
    <div className={styles.side}>
      {plant !== null && (
        <Outline
          plant={plant.plant}
          hall={null}
          selected={state.selected}
          onSelect={(id) => {
            review.select(id);
          }}
        />
      )}
      <Navigator />
    </div>
  );
}

export function RunScreen({ review, plants, workspace, navigation, worldId }: RunScreenProps) {
  const state = useStore(review);
  const plant = useRunPlant(state, plants, workspace);
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
        <Header state={state} navigation={navigation} worldId={worldId} />
        <Side review={review} state={state} plant={plant} />
        <section className={styles.center} aria-label={`Run ${worldId}`}>
          <Banners state={state} />
          <Timeline review={review} state={state} />
          <Graph review={review} state={state} plant={plant} />
        </section>
        <RunPanel review={review} state={state} plant={plant} unit={unit} />
      </div>
    </ReactFlowProvider>
  );
}
