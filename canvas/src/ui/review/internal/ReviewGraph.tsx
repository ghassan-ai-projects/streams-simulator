import { useMemo } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { readyReport, startNsOf } from "../../../app/review";
import { plantView } from "../../../model/graph";
import type { PlantView } from "../../../model/graph";
import { firstEventNode, noteLinesByEntity, reviewView } from "../../../model/review";
import type { Report } from "../../../model/review";
import { PlantGraph, useCanvasAspect } from "../../graph";
import { screen } from "../../primitives";
import { GraphBar } from "./GraphBar";
import type { HallProps } from "./RunHeader";
import { useLive } from "./useLive";
import parts from "./ReportParts.module.css";
import styles from "./ReviewScreen.module.css";

function OpeningPlant({ state }: { readonly state: ReviewState }) {
  const waiting = state.connection === "running" && state.sessionDomain === "";
  return (
    <p className={parts.muted} role="status">
      {waiting ? "Waiting for the session to name its plant…" : "Opening the plant this run used…"}
    </p>
  );
}

export interface GraphProps extends HallProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
}

function anchorOf(view: PlantView | null, report: Report | null): string | null {
  return view === null ? null : firstEventNode(report, new Set(view.nodes.map((node) => node.id)));
}

export function Graph({ review, state, plant, hall, onHall }: GraphProps) {
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
  if (view === null || state.report.phase === "loading") return <OpeningPlant state={state} />;
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
