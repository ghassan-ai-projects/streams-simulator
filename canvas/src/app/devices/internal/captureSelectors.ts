import { fileSafeName } from "../../../model/devices";
import type { Card, CaptureState, Reviewed } from "./captureState";

export type Way = "create" | "replace";

export function reviewedOf(state: CaptureState): Reviewed | null {
  return state.phase === "reviewing" || state.phase === "saving" || state.phase === "saved" ? state : null;
}

export function canWrite(card: Card, way: Way): boolean {
  const open = card.decision.kind === "open" || card.decision.kind === "failed";
  const wanted = way === "create" ? "new" : "differs";
  return open && card.kind.comparison.status === wanted && fileSafeName(card.kind.kind);
}

export function decisionCounts(cards: readonly Card[]): { readonly saved: number; readonly skipped: number } {
  const count = (kind: Card["decision"]["kind"]) => cards.filter((card) => card.decision.kind === kind).length;
  return { saved: count("saved"), skipped: count("skipped") };
}
