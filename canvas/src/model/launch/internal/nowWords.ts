import { clockOf } from "../../shared/duration";
import { humanise } from "../../shared/words";
import type { CallOutcome, PlannedFault } from "./launch";

export function outcomeWords(outcome: CallOutcome): string {
  const call = `${clockOf(outcome.atS)} · ${outcome.entity} · ${humanise(outcome.effector)}`;
  const verdict = outcome.accepted ? "accepted" : `command not accepted: ${outcome.reason}`;
  return outcome.immediate ? `${call}: ${verdict} · now` : `${call}: ${verdict}`;
}

function injectedWords(fault: PlannedFault): string {
  return `${clockOf(fault.atS)} · ${fault.entity} · fault injected: ${humanise(fault.fault)} · now`;
}

interface Line {
  readonly atS: number;
  readonly words: string;
}

export function nowLines(calls: readonly CallOutcome[], injected: readonly PlannedFault[]): readonly string[] {
  const lines: Line[] = [
    ...calls.filter((call) => call.immediate).map((call) => ({ atS: call.atS, words: outcomeWords(call) })),
    ...injected.map((fault) => ({ atS: fault.atS, words: injectedWords(fault) })),
  ];
  return lines.sort((a, b) => a.atS - b.atS).map((line) => line.words);
}
