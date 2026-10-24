import { useEffect, useRef } from "react";
import type { ReviewSession } from "../../../app/review";
import type { WorkspaceState, WorkspaceStore } from "../../../app/workspace";
import type { RunEntry } from "../../../model/workspace";
import { useStore } from "../../primitives";

function entryOf(listing: WorkspaceState, worldId: string): RunEntry | undefined {
  return listing.phase === "ready" ? listing.workspace.runs.find((run) => run.worldId === worldId) : undefined;
}

export type OpenMode = "review" | "live" | "watch";

export function useReviewOpen(
  review: ReviewSession,
  workspace: WorkspaceStore,
  worldId: string,
  mode: OpenMode = "review",
): void {
  const listing = useStore(workspace);
  const known = useRef<RunEntry | undefined>(undefined);
  const settled = listing.phase !== "loading";
  useEffect(() => {
    known.current = entryOf(listing, worldId);
    review.refreshCompleted(worldId, known.current);
  }, [listing, worldId, review]);
  useEffect(() => {
    if (!settled) return undefined;
    if (mode === "live") review.openLive(worldId, known.current);
    else if (mode === "watch") review.watch(worldId, known.current);
    else review.open(worldId, known.current);
    return () => {
      review.close();
    };
  }, [review, worldId, settled, mode]);
}
