import type { ReactNode } from "react";
import { useCallback, useMemo } from "react";
import type { ConnectSession } from "../../../app/launch";
import type { ReviewSession, ReviewState } from "../../../app/review";
import type { ReadyPlant } from "../../../app/plant";
import type { WorkspaceStore } from "../../../app/workspace";
import { plantView } from "../../../model/graph";
import type { InstanceOrigin, ViewPresentation } from "../../../model/graph";
import type { PlannedFault } from "../../../model/review";
import { deviceLine, deviceOf, noteLinesByEntity, readingWords, reviewView } from "../../../model/review";
import { readyReport } from "../../../app/review";
import type { NodeLive } from "../../graph";
import { useStore } from "../../primitives";
import { Interrupts } from "./Interrupts";
import { useLive } from "./useLive";
import { ObservedReadings } from "./ObservedReadings";
import { TrendsPanel } from "./TrendsPanel";
import { TrendsToggle } from "./TrendsToggle";
import { useReviewOpen } from "./useReviewOpen";
import { useShownOrigins } from "./useShownOrigins";

interface DeviceView {
  readonly line: string;
  readonly readings: readonly {
    readonly name: string;
    readonly valueText: string;
    readonly stateWords: string;
    readonly unitTitle: string;
  }[];
}

interface ConnectedSession {
  readonly kind: "session" | "watch";
  readonly live: (id: string) => NodeLive;
  readonly observed: (id: string) => ReactNode;
  readonly device: (id: string) => DeviceView | null;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly interrupts: ReactNode;
  readonly run: string;
  readonly view: ViewPresentation;
  readonly notes: ReadonlyMap<string, number>;
  readonly trends: (select: (entity: string) => void, planned: readonly PlannedFault[]) => ReactNode;
  readonly trendsToggle: ReactNode;
}

interface Props {
  readonly kind?: "session" | "watch";
  readonly review: ReviewSession;
  readonly plant: ReadyPlant | null;
  readonly run: string | null;
  readonly connect: ConnectSession;
  readonly now: () => number;
  readonly children: (session: ConnectedSession | null) => ReactNode;
}

export function RunWatchHold({
  review,
  workspace,
  run,
}: {
  readonly review: ReviewSession;
  readonly workspace: WorkspaceStore;
  readonly run: string;
}) {
  useReviewOpen(review, workspace, run, "watch");
  return null;
}

export function LiveSessionHold({
  review,
  workspace,
  run,
}: {
  readonly review: ReviewSession;
  readonly workspace: WorkspaceStore;
  readonly run: string;
}) {
  useReviewOpen(review, workspace, run, "live");
  return null;
}

function useDeviceView(state: ReviewState): (id: string) => DeviceView | null {
  const { live } = state;
  return useCallback(
    (id: string) => {
      const device = deviceOf(live, id);
      if (device === null) return null;
      const readings = device.readings.map((reading) => ({ name: reading.name, ...readingWords(reading) }));
      return { line: deviceLine(device), readings };
    },
    [live],
  );
}

export function ConnectedPlant({ kind = "session", review, plant, run, connect, now, children }: Props) {
  const stored = useStore(review);
  const state = useShownOrigins(stored, plant);
  const view = useMemo(() => (plant === null ? null : plantView(plant.plant, plant.specs, { open: null })), [plant]);
  const report = readyReport(state);
  const notes = useMemo(() => noteLinesByEntity(report), [report]);
  const live = useLive(review, state, plant, view);
  const device = useDeviceView(state);
  const observed = (id: string) => <ObservedReadings review={review} state={state} plant={plant} entity={id} />;
  const following = run !== null && stored.worldId === run;
  const watched = kind === "watch";
  const interrupts = watched ? null : <Interrupts review={review} state={state} connect={connect} now={now} />;
  const trends = (select: (entity: string) => void, planned: readonly PlannedFault[]) => (
    <TrendsPanel review={review} state={state} plant={plant} onPick={select} planned={planned} />
  );
  const trendsToggle = <TrendsToggle review={review} plant={plant} />;
  const session = {
    kind,
    live,
    observed,
    device,
    origins: state.origins,
    interrupts,
    run: run ?? "",
    view: reviewView,
    notes,
    trends,
    trendsToggle,
  };
  return <>{children(following ? session : null)}</>;
}
