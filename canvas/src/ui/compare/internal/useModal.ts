import { useEffect, useRef } from "react";

const focusable = "button:not(:disabled), [href], input:not(:disabled), select:not(:disabled), summary, [tabindex='0']";

function focusablesIn(node: HTMLElement): HTMLElement[] {
  return [...node.querySelectorAll<HTMLElement>(focusable)];
}

function wrapTab(event: KeyboardEvent, node: HTMLElement): void {
  const items = focusablesIn(node);
  const first = items[0];
  const last = items.at(-1);
  if (first === undefined || last === undefined) return;
  const active = document.activeElement;
  const loose = active === node || !(active instanceof HTMLElement && node.contains(active));
  const end = event.shiftKey ? first : last;
  if (active !== end && !loose) return;
  event.preventDefault();
  (event.shiftKey ? last : first).focus();
}

export function useModal<T extends HTMLElement>(onClose: () => void) {
  const ref = useRef<T>(null);
  const close = useRef(onClose);
  useEffect(() => {
    close.current = onClose;
  });
  useEffect(() => {
    const node = ref.current;
    if (node === null) return;
    const opener = document.activeElement;
    node.focus();
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.stopPropagation();
        close.current();
      } else if (event.key === "Tab") wrapTab(event, node);
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
      if (opener instanceof HTMLElement) opener.focus();
    };
  }, []);
  return ref;
}
