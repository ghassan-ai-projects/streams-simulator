import type { Plant } from "../../plant";
import { callArgs, effectorOf } from "../../plant";
import type { DraftCall } from "./draft";
import type { PlannedCall } from "./launch";
import type { CallProblem } from "./planProblems";
import { specFinding, targetFinding } from "./planProblems";

export type NowDraft = Pick<DraftCall, "entity" | "effector" | "values">;

export const emptyNowDraft: NowDraft = { entity: "", effector: "", values: {} };

type NowCall = Pick<PlannedCall, "entity" | "effector" | "args">;

export function nowCallOf(draft: NowDraft, plant: Plant): NowCall {
  const { entity, effector, values } = draft;
  return { entity, effector, args: callArgs(effectorOf(plant, entity, effector), entity, values) };
}

export function nowCallProblem(draft: NowDraft, plant: Plant): CallProblem | null {
  const call = nowCallOf(draft, plant);
  const name = "The call";
  const found =
    targetFinding(call, name) ?? specFinding(call, name, (entity, effector) => effectorOf(plant, entity, effector));
  return found === null ? null : { field: "call", index: 0, ...found };
}
