import { useEffect } from "react";
import type { EventMark } from "../../model/timeline";
import { nextEvent, previousEvent } from "../../model/timeline";
import { inTextField, plainKey } from "../keys";

export function useEventKeys(marks: readonly EventMark[], positionNs: number, onSeek: (ns: number) => void) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if ((event.key !== "," && event.key !== ".") || !plainKey(event) || inTextField(event.target)) return;
      const target = event.key === "," ? previousEvent(marks, positionNs) : nextEvent(marks, positionNs);
      if (target === null) return;
      event.preventDefault();
      onSeek(target);
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
    };
  }, [marks, positionNs, onSeek]);
}
