import { useEffect, useRef } from "react";

export function useFocusOnMount<T extends HTMLElement>(when: boolean | number = true) {
  const ref = useRef<T>(null);
  useEffect(() => {
    if (when === false || when === 0) return;
    ref.current?.focus();
    if (document.activeElement !== ref.current) {
      requestAnimationFrame(() => {
        ref.current?.focus();
      });
    }
  }, [when]);
  return ref;
}
