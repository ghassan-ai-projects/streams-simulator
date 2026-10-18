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

interface ConnectedSession {
  readonly live: (id: string) => NodeLive;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly run: string;
}

interface Props {
  readonly review: ReviewSession;
  readonly workspace: WorkspaceStore;
  readonly plant: ReadyPlant | null;
  readonly run: string | null;
  readonly children: (session: ConnectedSession | null) => ReactNode;
}

function FollowSession({ review, workspace, run }: Pick<Props, "review" | "workspace"> & { readonly run: string }) {
  useReviewOpen(review, workspace, run);
  return null;
}

export function ConnectedPlant({ review, workspace, plant, run, children }: Props) {
  const stored = useStore(review);
  const state = useShownOrigins(stored, plant);
  const view = useMemo(() => (plant === null ? null : plantView(plant.plant, plant.specs, { open: null })), [plant]);
  const live = useLive(review, state, plant, view);
  const following = run !== null && stored.worldId === run;
  return (
    <>
      {run !== null && <FollowSession review={review} workspace={workspace} run={run} />}
      {children(following ? { live, origins: state.origins, run } : null)}
    </>
  );
}
