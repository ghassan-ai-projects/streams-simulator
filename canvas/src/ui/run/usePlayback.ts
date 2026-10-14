import { useEffect } from "react";
import type { ReviewStore } from "../../app/reviewStore";
import { ownsSpace, plainKey } from "../keys";

const frameMs = 100;

export function usePlayback(review: ReviewStore, playing: boolean) {
  useEffect(() => {
    if (!playing) return undefined;
    let last = performance.now();
    const timer = setInterval(() => {
      const now = performance.now();
      review.tick(now - last);
      last = now;
    }, frameMs);
    return () => {
      clearInterval(timer);
    };
  }, [review, playing]);
  useEffect(
    () => () => {
      review.pause();
    },
    [review],
  );
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
