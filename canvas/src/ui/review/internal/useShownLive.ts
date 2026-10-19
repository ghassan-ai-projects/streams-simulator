import { useMemo } from "react";
import type { ReviewSession, ReviewState } from "../../../app/review";
import type { SessionLive } from "../../../model/review";
import { newestInterrupts, withShownInterrupts } from "../../../model/review";
import { useStore } from "../../primitives";

export function useShownLive(review: ReviewSession, state: ReviewState): SessionLive | null {
  const controls = useStore(review.controls);
  const { live, worldId } = state;
  const newest = state.liveHistory?.entries.at(-1) ?? null;
  const shown = newestInterrupts(newest, controls.seen, worldId);
  return useMemo(() => withShownInterrupts(live, newest, shown), [live, newest, shown]);
}
