import { clockOf } from "../../shared/duration";
import { rateWords } from "./launch";

interface RunProgress {
  readonly atS: number;
  readonly durationS: number;
  readonly rate: number | null;
  readonly paused: boolean;
}

export function runStageWords(progress: Pick<RunProgress, "paused">): string {
  return progress.paused ? "Paused" : "Running";
}

export function runProgressWords(progress: RunProgress, unit = ""): string {
  const speed = progress.paused ? "" : rateWords(progress.rate);
  return `${clockOf(progress.atS)} of ${clockOf(progress.durationS)}${unit}${speed === "" ? "" : ` · ${speed}`}`;
}
