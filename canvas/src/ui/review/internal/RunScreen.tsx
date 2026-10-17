import { shared, screen, cx } from "../../primitives";
import { ReactFlowProvider } from "@xyflow/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { PlantSession, ReadyPlant } from "../../../app/plant";
import type { ReviewPlant, ReviewState, ReviewSession } from "../../../app/review";
import { readyReport, reviewPlantFor, showRunPlant, startNsOf } from "../../../app/review";
import type { WorkspaceState, WorkspaceStore } from "../../../app/workspace";
import type { RunEntry } from "../../../model/workspace";
import { stateUnit } from "../../../model/plant";
import { noteLinesByEntity } from "../../../model/review";
import { plantView } from "../../../model/graph";
import { Navigator, Outline, PlantGraph, useCanvasAspect } from "../../graph";
import { reviewView } from "../../../model/review";
import type { PlantView } from "../../../model/graph";
import { firstEventNode } from "../../../model/review";
import type { Report } from "../../../model/review";
import { runSpan } from "../../../model/review";
import { useStore } from "../../primitives";
import { MissingRun } from "./MissingRun";
import { Banners } from "./RunStatus";
import type { HallProps } from "./RunHeader";
import { Header } from "./RunHeader";
import { RunPanel } from "./RunPanel";
import { Scrubber } from "./Scrubber";
import { GraphBar } from "./GraphBar";
import { useLive } from "./useLive";
import { usePlayback } from "./usePlayback";
import parts from "./ReportParts.module.css";
import styles from "./RunScreen.module.css";

interface RunScreenProps {
  readonly review: ReviewSession;
  readonly plants: PlantSession;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly worldId: string;
}

function entryOf(listing: WorkspaceState, worldId: string): RunEntry | undefined {
  return listing.phase === "ready" ? listing.workspace.runs.find((run) => run.worldId === worldId) : undefined;
}

function useReviewOpen(review: ReviewSession, workspace: WorkspaceStore, worldId: string): void {
  const listing = useStore(workspace);
  const known = useRef<RunEntry | undefined>(undefined);
  const settled = listing.phase !== "loading";
  useEffect(() => {
    known.current = entryOf(listing, worldId);
    review.refreshCompleted(worldId, known.current);
  }, [listing, worldId, review]);
  useEffect(() => {
    if (!settled) return undefined;
    review.open(worldId, known.current);
    return () => {
      review.close();
    };
  }, [review, worldId, settled]);
}

function useClosePlant(plants: PlantSession, worldId: string): void {
  useEffect(
    () => () => {
      plants.close();
    },
    [plants, worldId],
  );
}

function useRunPlant(state: ReviewState, plants: PlantSession, workspace: WorkspaceStore): ReviewPlant {
  const listing = useStore(workspace);
  const shown = useStore(plants);
  const known = listing.phase === "ready" ? listing.workspace : null;
  const view = reviewPlantFor(state, shown, known);
  const { runDomain } = state;
  const { key, fallback } = view;
  useEffect(() => {
    showRunPlant(plants, { key, fallback }, runDomain, known);
  }, [runDomain, key, fallback, known, plants]);
  return view;
}

function OpeningPlant() {
  return (
    <p className={parts.muted} role="status">
      Opening the plant this run used…
    </p>
  );
}

interface GraphProps extends HallProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
}

function anchorOf(view: PlantView | null, report: Report | null): string | null {
  return view === null ? null : firstEventNode(report, new Set(view.nodes.map((node) => node.id)));
}

function Graph({ review, state, plant, hall, onHall }: GraphProps) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  const report = readyReport(state);
  const notes = useMemo(() => noteLinesByEntity(report), [report]);
  const view = useMemo(
    () =>
      plant === null ? null : plantView(plant.plant, plant.specs, { open: hall, view: reviewView, aspect, notes }),
    [plant, hall, aspect, notes],
  );
  const live = useLive(review, state, plant, view);
  const anchor = useMemo(() => anchorOf(view, report), [view, report]);
  if (view === null) return <OpeningPlant />;
  return (
    <div className={styles.graphArea}>
      <GraphBar state={state} opening={review.index.opening()} startNs={startNsOf(state, review)} />
      <div ref={ref} className={screen.graphHost}>
        <PlantGraph
          anchor={anchor}
          panCue
          view={view}
          viewKey={hall ?? "review"}
          fitKey={`${hall ?? "review"}:${String(aspect)}`}
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
    </div>
  );
}

function Timeline({ review, state }: { readonly review: ReviewSession; readonly state: ReviewState }) {
  const report = readyReport(state);
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
  readonly review: ReviewSession;
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

function useUnitOf(plant: ReadyPlant | null) {
  return useCallback(
    (entity: string, name: string) => (plant === null ? "" : stateUnit(plant.plant, entity, name)),
    [plant],
  );
}

interface CentreProps extends GraphProps {
  readonly worldId: string;
  readonly navigation: NavigationStore;
  readonly review: ReviewSession;
}

function Centre({ review, state, plant, hall, onHall, worldId, navigation }: CentreProps) {
  return (
    <section className={styles.center} aria-label={`Run ${worldId}`}>
      <Banners state={state} />
      <Timeline review={review} state={state} />
      {state.report.phase === "failed" && state.report.failure.kind === "missing" ? (
        <MissingRun worldId={worldId} navigation={navigation} />
      ) : (
        <Graph review={review} state={state} plant={plant} hall={hall} onHall={onHall} />
      )}
    </section>
  );
}

export function RunScreen({ review, plants, workspace, navigation, launch, worldId }: RunScreenProps) {
  const state = useStore(review);
  const view = useRunPlant(state, plants, workspace);
  const { plant } = view;
  const [hall, setHall] = useState<string | null>(null);
  useReviewOpen(review, workspace, worldId);
  useClosePlant(plants, worldId);
  usePlayback(review, state.playing);
  const unit = useUnitOf(plant);
  return (
    <ReactFlowProvider>
      <div className={cx(styles.screen, shared.panels)}>
        <Header
          state={state}
          navigation={navigation}
          launch={launch}
          worldId={worldId}
          hall={hall}
          onHall={setHall}
          plant={view}
        />
        <Side review={review} state={state} plant={plant} hall={hall} onHall={setHall} />
        <Centre
          review={review}
          state={state}
          plant={plant}
          hall={hall}
          onHall={setHall}
          worldId={worldId}
          navigation={navigation}
        />
        <RunPanel review={review} state={state} plant={plant} unit={unit} />
      </div>
    </ReactFlowProvider>
  );
}
