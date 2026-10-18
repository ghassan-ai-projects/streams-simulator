import type { ReactNode } from "react";
import { useCallback, useMemo } from "react";
import type { ReviewSession, ReviewState } from "../../../app/review";
import type { ReadyPlant } from "../../../app/plant";
import type { WorkspaceStore } from "../../../app/workspace";
import { plantView } from "../../../model/graph";
import type { InstanceOrigin } from "../../../model/graph";
import { deviceLine, deviceOf, readingWords } from "../../../model/review";
import type { NodeLive } from "../../graph";
import { useStore } from "../../primitives";
import { useLive } from "./useLive";
import { useReviewOpen } from "./useReviewOpen";
import { useShownOrigins } from "./useShownOrigins";

interface DeviceView {
  readonly line: string;
  readonly readings: readonly { readonly name: string; readonly valueText: string; readonly stateWords: string }[];
}

interface ConnectedSession {
  readonly live: (id: string) => NodeLive;
  readonly device: (id: string) => DeviceView | null;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly run: string;
}

interface Props {
  readonly review: ReviewSession;
  readonly plant: ReadyPlant | null;
  readonly run: string | null;
  readonly children: (session: ConnectedSession | null) => ReactNode;
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

export function ConnectedPlant({ review, plant, run, children }: Props) {
  const stored = useStore(review);
  const state = useShownOrigins(stored, plant);
  const view = useMemo(() => (plant === null ? null : plantView(plant.plant, plant.specs, { open: null })), [plant]);
  const live = useLive(review, state, plant, view);
  const device = useDeviceView(state);
  const following = run !== null && stored.worldId === run;
  return <>{children(following ? { live, device, origins: state.origins, run } : null)}</>;
}
