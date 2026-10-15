import { useCallback, useEffect, useRef, useState } from "react";
import type { RefObject } from "react";

/** True while a scrolling element has more content below the cut. */
export function useCutBottom<T extends HTMLElement>(
  content: unknown,
): {
  readonly ref: RefObject<T | null>;
  readonly cut: boolean;
} {
  const ref = useRef<T | null>(null);
  const [cut, setCut] = useState(false);
  const measure = useCallback(() => {
    const node = ref.current;
    setCut(node !== null && node.scrollTop + node.clientHeight < node.scrollHeight - 1);
  }, []);
  useEffect(() => {
    const node = ref.current;
    if (node === null) return undefined;
    node.addEventListener("scroll", measure, { passive: true });
    const watcher = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(measure);
    watcher?.observe(node);
    return () => {
      node.removeEventListener("scroll", measure);
      watcher?.disconnect();
    };
  }, [measure]);
  useEffect(measure, [measure, content]);
  return { ref, cut };
}
