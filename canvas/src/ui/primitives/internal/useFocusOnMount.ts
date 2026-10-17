import { useEffect, useRef } from "react";

interface Options {
  readonly when: boolean;
}

export function useFocusOnMount<T extends HTMLElement>({ when }: Options = { when: true }) {
  const ref = useRef<T>(null);
  useEffect(() => {
    if (!when) return;
    ref.current?.focus();
    if (document.activeElement !== ref.current) {
      requestAnimationFrame(() => {
        ref.current?.focus();
      });
    }
  }, [when]);
  return ref;
}
