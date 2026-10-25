import type { PlanMark } from "./types";

const charsPerLine = 16;

type Notes = ReadonlyMap<string, number>;

function linesOf(text: string): number {
  return text === "" ? 0 : Math.ceil(text.length / charsPerLine);
}

export function noteLineCount(fault: string, effector: string): number {
  const cause = fault === "" ? "" : `Fault · ${fault}`;
  return Math.max(1, linesOf(cause) + linesOf(effector));
}

export function markNoteLines(marks: ReadonlyMap<string, PlanMark> | undefined): Notes {
  const found = new Map<string, number>();
  for (const [entity, mark] of marks ?? []) {
    if (mark.fault !== "" || mark.effector !== "") found.set(entity, noteLineCount(mark.fault, mark.effector));
  }
  return found;
}

export function widestNotes(first: Notes | undefined, second: Notes): Notes {
  const wide = new Map(first);
  for (const [entity, lines] of second) wide.set(entity, Math.max(lines, wide.get(entity) ?? 0));
  return wide;
}
