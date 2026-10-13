import { ReactFlowProvider } from "@xyflow/react";
import { ChevronRight, CircleCheck, CircleX, Eye, Lock, TriangleAlert } from "lucide-react";
import { useCallback, useEffect, useMemo } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantStore } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import { plantView } from "../../model/plantView";
import type { Report } from "../../model/report";
import { badgeLabels, badgeOf, runSpan } from "../../model/report";
import { readoutHeight } from "../plant/nodes";
import { Navigator, ZoomBar } from "../plant/GraphControls";
import { PlantGraph } from "../plant/PlantGraph";
import { useStore } from "../useStore";
import { ReportPanel } from "./ReportPanel";
import { Scrubber } from "./Scrubber";
import shared from "../shared.module.css";
import styles from "./RunScreen.module.css";

interface RunScreenProps {
  readonly review: ReviewStore;
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly worldId: string;
}

const badgeIcons = { verified: CircleCheck, "not-verified": CircleX, incomplete: TriangleAlert } as const;

function Badges({ state }: { readonly state: ReviewState }) {
  if (state.report.phase !== "ready") {
    return (
      <span className={styles.chip}>{state.report.phase === "loading" ? "Building the report…" : "No report"}</span>
    );
  }
  const report = state.report.value;
  const matches = state.verified.phase === "ready" ? state.verified.value : null;
  const badge = badgeOf(report, matches);
  const Icon = badgeIcons[badge];
  return (
    <>
      <span className={styles[badge]}>
        <Icon size={13} aria-hidden="true" />
        {state.verified.phase === "loading" && badge === "verified" ? "Verifying…" : badgeLabels[badge]}
      </span>
      <span className={report.mode === "explore" ? styles.truth : styles.chip}>
        {report.mode === "explore" ? <Eye size={13} aria-hidden="true" /> : <Lock size={13} aria-hidden="true" />}
        {report.mode === "explore" ? "Explore · unblinded" : "Evaluate · ended"}
      </span>
    </>
  );
}

function Header({
  state,
  navigation,
  worldId,
}: {
  readonly state: ReviewState;
  readonly navigation: NavigationStore;
  readonly worldId: string;
}) {
  const progress =
    state.playback.phase === "loading"
      ? `loading observations · ${state.playback.records.toLocaleString("en-US")}`
      : `${state.playback.records.toLocaleString("en-US")} observations`;
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
      <h1 className={styles.here}>{worldId}</h1>
      <Badges state={state} />
      <span className={styles.progress} role="status">
        {progress}
      </span>
      <ZoomBar />
    </div>
  );
}

function usePlant(state: ReviewState, plants: PlantStore, workspace: WorkspaceStore) {
  const listing = useStore(workspace);
  const plant = useStore(plants);
  const domainId = state.report.phase === "ready" ? state.report.value.domainId : null;
  const path =
    listing.phase === "ready" ? listing.workspace.domains.find((entry) => entry.id === domainId)?.path : undefined;
  useEffect(() => {
    if (path !== undefined && listing.phase === "ready") void plants.load(path, listing.workspace);
  }, [path, listing, plants]);
  return plant.phase === "ready" && plant.path === path ? plant : null;
}

function Graph({
  review,
  state,
  plants,
  workspace,
}: {
  readonly review: ReviewStore;
  readonly state: ReviewState;
  readonly plants: PlantStore;
  readonly workspace: WorkspaceStore;
}) {
  const plant = usePlant(state, plants, workspace);
  const view = useMemo(
    () => (plant === null ? null : plantView(plant.plant, plant.specs, null, readoutHeight)),
    [plant],
  );
  const readings = useCallback((id: string) => review.index.readings(id, state.positionNs), [review, state.positionNs]);
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
      readings={readings}
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

export function RunScreen({ review, plants, workspace, navigation, worldId }: RunScreenProps) {
  const state = useStore(review);
  useEffect(() => {
    review.open(worldId);
  }, [review, worldId]);
  return (
    <ReactFlowProvider>
      <div className={styles.screen}>
        <Header state={state} navigation={navigation} worldId={worldId} />
        <div className={styles.side}>
          <Navigator />
        </div>
        <section className={styles.center} aria-label={`Run ${worldId}`}>
          <Timeline review={review} state={state} />
          <Graph review={review} state={state} plants={plants} workspace={workspace} />
        </section>
        <aside className={styles.report} aria-label="Report" tabIndex={0}>
          {state.report.phase === "ready" && <ReportPanel report={state.report.value} />}
          {state.report.phase === "loading" && (
            <p className={styles.muted} role="status">
              streamsim report is reading the run folder…
            </p>
          )}
          {state.report.phase === "failed" && (
            <p
              className={styles.failure}
              role="alert"
            >{`${state.report.failure.message} (${state.report.failure.code})`}</p>
          )}
        </aside>
      </div>
    </ReactFlowProvider>
  );
}
