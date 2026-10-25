import { durationOf } from "../../shared/duration";
import type { RunEntry } from "./types";

const stoppedByOperator = "stopped_by_operator";
const singleUnit = /^\d+ (s|min|h)$/;

function unitOf(words: string): string {
  return words.slice(words.lastIndexOf(" ") + 1);
}

function reachedOf(reached: string, asked: string): string {
  return singleUnit.test(reached) && singleUnit.test(asked) && unitOf(reached) === unitOf(asked)
    ? reached.slice(0, reached.lastIndexOf(" "))
    : reached;
}

function stoppedEarly(run: RunEntry): boolean {
  return run.terminationReason === stoppedByOperator && (run.requestedS ?? 0) > 0;
}

export function lengthFacts(run: RunEntry): readonly string[] {
  if (!run.ended) return [];
  if (!stoppedEarly(run)) return [durationOf(run.durationS)];
  const asked = durationOf(run.requestedS ?? 0);
  return [`${reachedOf(durationOf(run.durationS) || "0 s", asked)} of ${asked}`, "stopped"];
}
