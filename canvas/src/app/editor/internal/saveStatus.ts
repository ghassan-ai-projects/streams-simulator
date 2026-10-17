import { issuesWords } from "../../../model/editor";
import type { Draft } from "./editorDraft";
import { checked, emptyPlant, isDirty } from "./editorSelectors";
import type { EditorFailure } from "../../ports";

export interface SaveStatus {
  readonly text: string;
  readonly detail: string | null;
  readonly tone: "busy" | "warn" | "ok" | "";
}

function status(text: string, tone: SaveStatus["tone"], detail: string | null = null): SaveStatus {
  return { text, detail, tone };
}

function composeFailed(draft: Draft, failure: EditorFailure): SaveStatus {
  const listed = draft.issues?.listed.length ?? 0;
  if (listed > 0)
    return status(`Needs fixing: ${issuesWords(listed)}, listed in Build. Your changes are saved.`, "warn");
  return status(`Saved, but composition failed: ${failure.diagnostic.message}`, "warn");
}

function settled(draft: Draft): SaveStatus {
  if (checked(draft)) return status("No changes", "");
  const text = emptyPlant(draft)
    ? "Nothing to save yet: add equipment in Build."
    : "Not checked yet: Save to see its issues.";
  return status(text, "");
}

export function saveStatusOf(draft: Draft): SaveStatus {
  const { save } = draft;
  if (save.kind === "saving") return status("Saving and composing…", "busy");
  if (save.kind === "write-failed") return status(`Not saved: ${save.failure.diagnostic.message}`, "warn");
  if (isDirty(draft)) return status("Unsaved changes", "busy");
  if (save.kind === "compose-failed") return composeFailed(draft, save.failure);
  if (save.kind === "composed") return status("Ready to run. Saved.", "ok", save.summary);
  return settled(draft);
}
