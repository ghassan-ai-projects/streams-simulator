import { useCallback, useState } from "react";
import type { KeyboardEvent } from "react";
import type { PlaybackIndex, Plot, Span } from "../../../model/review";
import { probeKey } from "../../../model/review";

export interface Probe {
  readonly ns: number | null;
  readonly clear: () => void;
  readonly move: (ns: number) => void;
  readonly onKey: (event: KeyboardEvent, plot: Plot, at: number) => "commit" | null;
}

export function useProbe(index: PlaybackIndex, span: Span): Probe {
  const [ns, move] = useState<number | null>(null);
  const clear = useCallback(() => {
    move(null);
  }, []);
  const onKey = useCallback(
    (event: KeyboardEvent, plot: Plot, at: number): "commit" | null => {
      const watched = { entity: plot.entity, channel: plot.channel };
      const step = probeKey(index, watched, span, at, { key: event.key, shift: event.shiftKey });
      if (step === null) return null;
      event.preventDefault();
      event.stopPropagation();
      if (step.kind === "commit") return "commit";
      move(step.kind === "move" ? step.atNs : null);
      return null;
    },
    [index, span],
  );
  return { ns, clear, move, onKey };
}
