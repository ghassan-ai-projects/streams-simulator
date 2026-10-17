import { useEffect, useRef, useState } from "react";

export function useCanvasAspect<T extends HTMLElement>() {
  const ref = useRef<T>(null);
  const [aspect, setAspect] = useState<number | undefined>(undefined);
  useEffect(() => {
    const node = ref.current;
    if (node === null || typeof ResizeObserver === "undefined") return undefined;
    const observer = new ResizeObserver(() => {
      const { width, height } = node.getBoundingClientRect();
      if (width > 0 && height > 0) setAspect(Math.round((width / height) * 4) / 4);
    });
    observer.observe(node);
    return () => {
      observer.disconnect();
    };
  }, []);
  return { ref, aspect };
}
