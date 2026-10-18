import type { ReactNode } from "react";
import { useMemo } from "react";
import type { ReviewSession } from "../../../app/review";
import type { ReadyPlant } from "../../../app/plant";
import type { WorkspaceStore } from "../../../app/workspace";
import { plantView } from "../../../model/graph";
import type { InstanceOrigin } from "../../../model/graph";
import type { NodeLive } from "../../graph";
import { useStore } from "../../primitives";
import { useLive } from "./useLive";
import { useReviewOpen } from "./useReviewOpen";
import { useShownOrigins } from "./useShownOrigins";

interface Props {
  readonly review: ReviewSession;
  readonly workspace: WorkspaceStore;
  readonly plant: ReadyPlant | null;
  readonly run: string;
  readonly children: (session: {
    readonly live: (id: string) => NodeLive;
    readonly origins: ReadonlyMap<string, InstanceOrigin>;
    readonly run: string;
  }) => ReactNode;
}

export function ConnectedPlant({ review, workspace, plant, run, children }: Props) {
  useReviewOpen(review, workspace, run);
  const stored = useStore(review);
  const state = useShownOrigins(stored, plant);
  const view = useMemo(() => (plant === null ? null : plantView(plant.plant, plant.specs, { open: null })), [plant]);
  const live = useLive(review, state, plant, view);
  return children({ live, origins: state.origins, run });
}
