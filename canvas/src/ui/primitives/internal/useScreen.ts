import { useEffect, useRef } from "react";

export function useScreen(title: string) {
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    document.title = `${title} · Streams Simulator`;
    const holding = document.activeElement?.closest("[role='application'], [role='dialog'], dialog") ?? null;
    if (holding === null) heading.current?.focus({ preventScroll: true });
  }, [title]);
  return heading;
}
