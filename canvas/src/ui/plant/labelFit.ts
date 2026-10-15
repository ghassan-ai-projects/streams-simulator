import type { RefObject } from "react";
import { useLayoutEffect } from "react";

export function useLabelFit(ref: RefObject<HTMLElement | null>, zoom: number): void {
  useLayoutEffect(() => {
    const element = ref.current;
    if (element === null) return;
    element.style.visibility = element.scrollWidth > element.clientWidth + 1 ? "hidden" : "";
  }, [ref, zoom]);
}
