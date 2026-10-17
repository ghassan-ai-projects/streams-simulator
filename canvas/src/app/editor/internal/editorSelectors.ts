import { domainOf, equipmentOf } from "../../../model/editor";
import type { Draft, EditorState } from "./editorDraft";

export function isDirty(state: EditorState): boolean {
  return state.phase === "ready" && state.doc !== state.saved;
}

export function isSaving(state: EditorState): boolean {
  return state.phase === "ready" && state.save.kind === "saving";
}

export function emptyPlant(draft: Draft): boolean {
  return equipmentOf(draft.doc).length === 0;
}

export function checked(draft: Draft): boolean {
  if (isDirty(draft)) return false;
  const known = draft.domain === domainOf(draft.doc);
  const { save } = draft;
  if (save.kind === "composed") return known || save.doc === draft.doc;
  return save.kind === "idle" && known;
}

export function canSave(state: EditorState): boolean {
  if (state.phase !== "ready") return false;
  const retry = state.save.kind === "compose-failed" || state.save.kind === "write-failed";
  return isDirty(state) || retry || (!checked(state) && !emptyPlant(state));
}
