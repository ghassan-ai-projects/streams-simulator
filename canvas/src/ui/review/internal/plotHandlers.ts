import type { Plot } from "../../../model/review";
import type { PlotHandlers } from "./TrendPlotSvg";
import type { Probe } from "./useProbe";

export function plotHandlers(
  probe: Probe,
  at: number,
  onPick: (entity: string, atNs: number) => void,
): (plot: Plot) => PlotHandlers {
  return (plot) => ({
    onKey: (event) => {
      if (probe.onKey(event, plot, at) === "commit") onPick(plot.entity, at);
    },
    onPoint: probe.move,
    onLeave: probe.clear,
    onPick: (ns) => {
      onPick(plot.entity, ns);
    },
    onBlur: probe.clear,
  });
}
