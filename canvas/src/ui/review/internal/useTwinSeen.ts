import { useEffect } from "react";
import type { ReviewSession } from "../../../app/review";

export function useTwinSeen(review: ReviewSession, shown: { readonly fresh: boolean }): void {
  const { fresh } = shown;
  useEffect(() => {
    if (fresh) review.twin.markSeen();
  }, [fresh, review]);
}
