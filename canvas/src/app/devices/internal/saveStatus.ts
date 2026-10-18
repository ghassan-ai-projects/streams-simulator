import { draftProblems, kindFacts } from "../../../model/devices";
import { counted } from "../../../model/workspace";
import type { DeviceFailure, KindChecked } from "../../ports";
import type { Draft } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";

export interface SaveStatus {
  readonly text: string;
  readonly detail: string | null;
  readonly tone: "busy" | "warn" | "ok" | "";
}

function status(text: string, tone: SaveStatus["tone"], detail: string | null = null): SaveStatus {
  return { text, detail, tone };
}

function writeFailed(failure: DeviceFailure): SaveStatus {
  if (failure.kind === "conflict") {
    return status("Not saved: the file changed on disk. Reload it to continue.", "warn", failure.diagnostic.message);
  }
  return status(`Not saved: ${failure.diagnostic.message}`, "warn");
}

function unsaved(draft: Draft): SaveStatus {
  const blocked = draftProblems(draft.doc).length > 0;
  return status(blocked ? "Unsaved changes. Name everything to save." : "Unsaved changes", "busy");
}

function refusedCount(draft: Draft, failure: DeviceFailure): string {
  const issues = draft.review.schema.length;
  return issues > 0 ? counted(issues, "issue") : failure.diagnostic.message;
}

function refused(draft: Draft, failure: DeviceFailure): SaveStatus {
  return status(`Saved, but the binary refuses it: ${refusedCount(draft, failure)}.`, "warn");
}

function notSavedRefused(draft: Draft, failure: DeviceFailure): SaveStatus {
  return status(`Not saved: the binary refuses it (${refusedCount(draft, failure)}). Fix it to save.`, "warn");
}

function checked(draft: Draft): SaveStatus {
  const { check } = draft;
  if (check.kind === "checking") return status("Checking with the binary…", "busy");
  if (check.kind === "valid") return valid(check.summary);
  if (check.kind === "refused") return refused(draft, check.failure);
  if (check.kind === "failed") return unchecked(check.failure);
  return status("Not checked yet.", "");
}

function valid({ metrics, commands }: KindChecked): SaveStatus {
  return status("Saved and valid.", "ok", kindFacts(metrics, commands));
}

function unchecked(failure: DeviceFailure): SaveStatus {
  return status(`Saved, but it could not be checked: ${failure.diagnostic.message}`, "warn");
}

export function saveStatusOf(draft: Draft): SaveStatus {
  const { save } = draft;
  if (save.kind === "saving") return status("Saving and checking…", "busy");
  if (save.kind === "write-failed") return writeFailed(save.failure);
  if (save.kind === "refused") return notSavedRefused(draft, save.failure);
  return isDirty(draft) ? unsaved(draft) : checked(draft);
}
