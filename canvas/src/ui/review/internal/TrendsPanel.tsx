import { useRef, useState } from "react";
import type { RefObject } from "react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { readoutOf, runClock } from "../../../model/review";
import type { PlannedFault } from "../../../model/review";
import { useResizeObserver, useStore } from "../../primitives";
import { TrendRow } from "./TrendRow";
import { EventLane, EventList, TimeAxis } from "./TrendLane";
import { plotHandlers } from "./plotHandlers";
import { TrendsHeader } from "./TrendsHeader";
import { TrendsHelp } from "./TrendsHelp";
import type { Probe } from "./useProbe";
import { useProbe } from "./useProbe";
import type { TrendsView } from "./useTrends";
import { useTrends } from "./useTrends";
import lane from "./TrendPlot.module.css";
import styles from "./Trends.module.css";

export interface TrendsPanelProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly onPick: (entity: string, atNs: number) => void;
  readonly planned?: readonly PlannedFault[];
}

const noFaults: readonly PlannedFault[] = [];
const defaultWidth = 600;
const step = 8;

function useTrackWidth() {
  const ref = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(defaultWidth);
  useResizeObserver(ref, () => {
    const measured = ref.current?.getBoundingClientRect().width ?? 0;
    if (measured > 0) setWidth(Math.max(160, Math.round(measured / step) * step));
  });
  return [ref, width] as const;
}

function Limited({ fromNs, startNs }: { readonly fromNs: number | null; readonly startNs: number }) {
  if (fromNs === null) return null;
  return (
    <p className={styles.limited} data-part="limited">
      {`Window limited: readings before ${runClock(fromNs, startNs)} are no longer kept while the run is watched, so the plots start there.`}
    </p>
  );
}

export function TrendsPanel(props: TrendsPanelProps) {
  const open = useStore(props.review.watchList).open;
  return open ? <TrendsBody {...props} /> : null;
}

function cursorWords(probeNs: number | null, cursorNs: number, startNs: number, running: boolean): string {
  if (probeNs !== null) return `Probe ${runClock(probeNs, startNs)}`;
  return `${running ? "Run clock" : "Cursor"} ${runClock(cursorNs, startNs)}`;
}

function TrendsBody(props: TrendsPanelProps) {
  const { review, state, plant } = props;
  const [trackRef, width] = useTrackWidth();
  const view = useTrends(review, state, plant, width, props.planned ?? noFaults);
  const probe = useProbe(review.index, view.span);
  return (
    <section className={styles.panel} aria-label="Trends">
      <TrendsHeader
        watch={review.watchList}
        plantId={plant?.plant.id ?? ""}
        count={view.watched.length}
        choice={view.choice}
        running={view.running}
        cursorWords={cursorWords(probe.ns, view.cursorNs, view.startNs, view.running)}
      />
      <Limited fromNs={view.limitedFromNs} startNs={view.startNs} />
      <div className={styles.scroll}>
        {view.watched.length === 0 ? (
          <p className={styles.empty}>
            Nothing watched yet. Select an instance and use the pin beside a channel in its Observed table to add it
            here.
          </p>
        ) : (
          <TrendsPlots {...props} view={view} probe={probe} width={width} trackRef={trackRef} />
        )}
        <TrendsHelp />
      </div>
    </section>
  );
}

interface PlotsProps extends TrendsPanelProps {
  readonly view: TrendsView;
  readonly probe: Probe;
  readonly width: number;
  readonly trackRef: RefObject<HTMLDivElement | null>;
}

function PlotRows({ review, plant, onPick, view, probe, width }: Omit<PlotsProps, "trackRef">) {
  const at = probe.ns ?? view.cursorNs;
  const marks = view.events.map((event) => event.atNs);
  const cursor = { span: view.span, width, cursorNs: view.cursorNs, probeNs: probe.ns, marks };
  const facts = { index: review.index, plant: plant?.plant ?? null, atNs: at, referenceNs: at, startNs: view.startNs };
  const handlers = plotHandlers(probe, at, onPick);
  return (
    <ol className={styles.rows} aria-label="Watched channels">
      {view.plots.map((plot) => (
        <TrendRow
          key={plot.key}
          plot={plot}
          cursor={cursor}
          readout={readoutOf(facts, plot)}
          handlers={handlers(plot)}
          onRemove={() => {
            review.watchList.toggle(plant?.plant.id ?? "", plot);
          }}
        />
      ))}
    </ol>
  );
}

function TrendsPlots(props: PlotsProps) {
  const { onPick, view, width, trackRef } = props;
  return (
    <>
      <div className={lane.lane}>
        <span>{view.events.length === 0 ? "No events" : "Events"}</span>
        <div ref={trackRef}>
          <EventLane events={view.events} span={view.span} width={width} />
        </div>
      </div>
      <PlotRows {...props} />
      <div className={lane.axis}>
        <span />
        <TimeAxis span={view.span} startNs={view.startNs} />
      </div>
      <EventList
        events={view.events}
        startNs={view.startNs}
        onPick={(event) => {
          onPick(event.entity, event.atNs);
        }}
      />
    </>
  );
}
