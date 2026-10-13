import { useEffect, useRef } from "react";

export function useScreen(title: string) {
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    document.title = `${title} · Streams Simulator`;
    heading.current?.focus({ preventScroll: true });
  }, [title]);
  return heading;
}
