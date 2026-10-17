import { useEffect } from "react";
import type { ReviewSession } from "../../../app/review";
import { ownsSpace, plainKey } from "../../primitives";

export function usePlaybackKeys(review: ReviewSession) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key !== " " || event.repeat || !plainKey(event) || ownsSpace(event.target)) return;
      event.preventDefault();
      if (review.snapshot().playing) review.pause();
      else review.play();
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
    };
  }, [review]);
}
