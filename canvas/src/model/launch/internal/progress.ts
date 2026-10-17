import { clockOf } from "../../shared/duration";
import { rateWords } from "./launch";

interface RunProgress {
  readonly atS: number;
  readonly durationS: number;
  readonly rate: number | null;
}

export function runProgressWords(progress: RunProgress, unit = ""): string {
  const speed = rateWords(progress.rate);
  return `${clockOf(progress.atS)} of ${clockOf(progress.durationS)}${unit}${speed === "" ? "" : ` · ${speed}`}`;
}
