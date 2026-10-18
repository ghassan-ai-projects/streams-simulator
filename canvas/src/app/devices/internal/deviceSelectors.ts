import { draftProblems, serialiseKind } from "../../../model/devices";
import type { FieldIssue } from "../../../model/devices";
import type { DeviceState, Draft } from "./deviceDraft";

export function isDirty(state: DeviceState): boolean {
  if (state.phase !== "ready" || state.doc === state.saved) return false;
  return serialiseKind(state.doc) !== serialiseKind(state.saved);
}

export function isSaving(state: DeviceState): boolean {
  return state.phase === "ready" && state.save.kind === "saving";
}

export function isBusy(state: DeviceState): boolean {
  return state.phase === "ready" && (state.save.kind === "saving" || state.check.kind === "checking");
}

export function canSave(state: DeviceState): boolean {
  if (state.phase !== "ready" || isSaving(state)) return false;
  if (state.save.kind === "write-failed" && state.save.failure.kind === "conflict") return false;
  return isDirty(state) && draftProblems(state.doc).length === 0;
}

export function interfaceIssues(draft: Draft): readonly FieldIssue[] {
  return [...draftProblems(draft.doc), ...draft.review.schema];
}

export function twinNeedsOf(draft: Draft): readonly FieldIssue[] {
  return draft.review.needs;
}
