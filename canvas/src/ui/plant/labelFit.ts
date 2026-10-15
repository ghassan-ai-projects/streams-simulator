import type { RefObject } from "react";
import { useLayoutEffect } from "react";

function overflows(element: HTMLElement): boolean {
  return element.scrollWidth > element.clientWidth + 1;
}

export function useLabelFit(ref: RefObject<HTMLElement | null>, zoom: number): void {
  useLayoutEffect(() => {
    const element = ref.current;
    if (element === null) return;
    element.style.visibility = "";
    element.dataset.form = "full";
    if (!overflows(element)) return;
    element.dataset.form = "short";
    if (overflows(element)) element.style.visibility = "hidden";
  }, [ref, zoom]);
}
