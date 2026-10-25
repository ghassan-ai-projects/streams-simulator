import { noteLineCount } from "../../graph";
import type { Report } from "./report";
import { callWords, faultNames } from "./report";

function faultsOf(report: Report): Map<string, string> {
  return new Map([...faultNames(report)].map(([entity, names]) => [entity, names.join(", ")] as const));
}

function callsOf(report: Report): Map<string, string> {
  const longest = new Map<string, string>();
  for (const call of report.effectorCalls) {
    const words = callWords(call);
    if (words.length > (longest.get(call.entity) ?? "").length) longest.set(call.entity, words);
  }
  return longest;
}

export function noteLinesByEntity(report: Report | null): ReadonlyMap<string, number> {
  if (report === null) return new Map();
  const faults = faultsOf(report);
  const calls = callsOf(report);
  const entities = new Set([...faults.keys(), ...calls.keys()]);
  return new Map(
    [...entities].map((entity) => [entity, noteLineCount(faults.get(entity) ?? "", calls.get(entity) ?? "")]),
  );
}
