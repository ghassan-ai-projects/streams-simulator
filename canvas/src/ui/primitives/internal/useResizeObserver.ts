import { useEffect, useRef } from "react";
import type { RefObject } from "react";

export function useResizeObserver(ref: RefObject<Element | null>, callback: () => void): void {
  const latest = useRef(callback);
  useEffect(() => {
    latest.current = callback;
  });
  useEffect(() => {
    const node = ref.current;
    if (node === null || typeof ResizeObserver === "undefined") return undefined;
    const observer = new ResizeObserver(() => {
      latest.current();
    });
    observer.observe(node);
    return () => {
      observer.disconnect();
    };
  }, [ref]);
}
