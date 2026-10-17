import { useEffect, useRef } from "react";

export function useInterval(callback: () => void, ms: number | null): void {
  const latest = useRef(callback);
  useEffect(() => {
    latest.current = callback;
  });
  useEffect(() => {
    if (ms === null) return undefined;
    const timer = setInterval(() => {
      latest.current();
    }, ms);
    return () => {
      clearInterval(timer);
    };
  }, [ms]);
}
