import { useCallback, useEffect, useRef, useState } from "react";
import type { RefObject } from "react";
import { useResizeObserver } from "./useResizeObserver";

/** True while a scrolling element has more content below the cut. */
export function useCutBottom<T extends HTMLElement>(): {
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
    const changes = new MutationObserver(measure);
    changes.observe(node, {
      subtree: true,
      childList: true,
      characterData: true,
      attributes: true,
      attributeFilter: ["open"],
    });
    return () => {
      node.removeEventListener("scroll", measure);
      changes.disconnect();
    };
  }, [measure]);
  useResizeObserver(ref, measure);
  useEffect(measure, [measure]);
  return { ref, cut };
}
