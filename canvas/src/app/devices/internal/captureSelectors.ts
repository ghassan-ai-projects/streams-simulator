import type { Card, CaptureState, Reviewed } from "./captureState";

export type Way = "create" | "replace";

export function reviewedOf(state: CaptureState): Reviewed | null {
  return state.phase === "reviewing" || state.phase === "saving" || state.phase === "saved" ? state : null;
}

export function savedEtag(card: Card): string | null {
  return card.saved?.ok === true ? card.saved.value.etag : null;
}

function actionable(card: Card): boolean {
  return card.decision.kind === "open" || card.decision.kind === "failed";
}

export function canWrite(card: Card, way: Way): boolean {
  const { status } = card.kind.comparison;
  if (!actionable(card)) return false;
  return way === "create" ? status === "new" : status === "differs" && savedEtag(card) !== null;
}

export function canSkip(card: Card): boolean {
  return actionable(card) && card.kind.comparison.status !== "same";
}

export function decisionCounts(cards: readonly Card[]): { readonly saved: number; readonly skipped: number } {
  const count = (kind: Card["decision"]["kind"]) => cards.filter((card) => card.decision.kind === kind).length;
  return { saved: count("saved"), skipped: count("skipped") };
}
