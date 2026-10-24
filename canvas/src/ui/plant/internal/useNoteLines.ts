import { useMemo } from "react";
import { markNoteLines, widestNotes } from "../../../model/graph";
import type { PlanMark } from "../../../model/graph";

type Notes = ReadonlyMap<string, number>;

export function useNoteLines(reported: Notes | undefined, marks: ReadonlyMap<string, PlanMark> | undefined): Notes {
  const signature = JSON.stringify([...widestNotes(reported, markNoteLines(marks))]);
  return useMemo(() => new Map(JSON.parse(signature) as [string, number][]), [signature]);
}
