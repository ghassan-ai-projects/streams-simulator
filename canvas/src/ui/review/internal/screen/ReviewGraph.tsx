import { useMemo } from "react";
import type { ReadyPlant } from "../../../../app/plant";
import type { ReviewSession, ReviewState } from "../../../../app/review";
import { readyReport, startNsOf } from "../../../../app/review";
import { plantView } from "../../../../model/graph";
import type { PlantView } from "../../../../model/graph";
import { firstEventNode, noteLinesByEntity, reviewView } from "../../../../model/review";
import type { Report } from "../../../../model/review";
import { PlantGraph, useCanvasAspect, useWires } from "../../../graph";
import { screen } from "../../../primitives";
import { GraphBar } from "./GraphBar";
import { TrendsToggle } from "../trends/TrendsToggle";
import type { HallProps } from "./RunHeader";
import { useLive } from "./useLive";
import parts from "../report/ReportParts.module.css";
import styles from "../ReviewScreen.module.css";

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

type WiresControl = ReturnType<typeof useWires>["control"];

function Bar({
  review,
  state,
  plant,
  wires,
}: Pick<GraphProps, "review" | "state" | "plant"> & { readonly wires: WiresControl }) {
  return (
    <GraphBar
      state={state}
      wires={wires}
      opening={review.index.opening()}
      startNs={startNsOf(state, review)}
      trends={<TrendsToggle review={review} plant={plant} />}
    />
  );
}

export function Graph({ review, state, plant, hall, onHall }: GraphProps) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  const report = readyReport(state);
  const notes = useMemo(() => noteLinesByEntity(report), [report]);
  const { wires, control } = useWires(plant?.plant ?? null);
  const view = useMemo(
    () =>
      plant === null
        ? null
        : plantView(plant.plant, plant.specs, { open: hall, view: reviewView, aspect, notes, wires }),
    [plant, hall, aspect, notes, wires],
  );
  const live = useLive(review, state, plant, view);
  const anchor = useMemo(() => anchorOf(view, report), [view, report]);
  if (view === null || state.report.phase === "loading") return <OpeningPlant state={state} />;
  return (
    <div className={styles.graphArea}>
      <Bar review={review} state={state} plant={plant} wires={hall === null ? control : null} />
      <div ref={ref} className={screen.graphHost}>
        <PlantGraph
          anchor={anchor}
          panCue
          view={view}
          viewKey={hall ?? "review"}
          fitKey={`${hall ?? "review"}:${String(aspect)}:${view.wires}`}
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
