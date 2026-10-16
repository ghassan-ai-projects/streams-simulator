import type { ShowingMarks } from "./shared/marks";
import { hallOf } from "./plantView";

export interface PlanMark {
  readonly status: "fault" | null;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}

function hallsOf(faulted: ReadonlyMap<string, readonly string[]>): Map<string, PlanMark> {
  const halls = new Map<string, number>();
  for (const entity of faulted.keys()) {
    const hall = hallOf(entity);
    if (hall !== null) halls.set(hall, (halls.get(hall) ?? 0) + 1);
  }
  const marks = new Map<string, PlanMark>();
  for (const [hall, count] of halls) marks.set(hall, { status: "fault", count, fault: "", effector: "" });
  return marks;
}

export function planMarks(marks: ShowingMarks): ReadonlyMap<string, PlanMark> {
  const faulted = marks.faults;
  const effectors = marks.effectors;
  const found = new Map<string, PlanMark>(hallsOf(faulted));
  for (const entity of new Set([...faulted.keys(), ...effectors.keys()])) {
    const causes = faulted.get(entity) ?? [];
    const status = causes.length > 0 ? "fault" : null;
    found.set(entity, { status, count: 0, fault: causes.join(", "), effector: effectors.get(entity) ?? "" });
  }
  return found;
}
