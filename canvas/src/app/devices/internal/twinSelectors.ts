import { choiceProblems, serialiseChoices } from "../../../model/devices";
import type { KindChoice, KindFile, TwinIssue } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { Built, TwinDraft, TwinState } from "./twinDraft";
import { interfaceNeedsOf } from "./twinNeeds";

export function isChoicesDirty(draft: TwinDraft): boolean {
  return serialiseChoices(draft.doc, draft.members) !== draft.written;
}

export function hasUnsavedChoices(state: TwinState): boolean {
  return state.phase === "ready" && state.edited && isChoicesDirty(state);
}

export function isStale(draft: TwinDraft): boolean {
  return draft.build.kind === "built" && serialiseChoices(draft.doc, draft.members) !== draft.build.fingerprint;
}

export function isBuilding(draft: TwinDraft): boolean {
  return draft.build.kind === "building";
}

export function builtOf(draft: TwinDraft): Built | null {
  return draft.build.kind === "built" && !isStale(draft) ? draft.build : null;
}

export function verdictOf(built: Built): "pending" | "failed" | "ready" {
  const { implemented, lint } = built;
  if (implemented.kind === "failed" || (lint.kind === "done" && lint.report.errors > 0)) return "failed";
  return implemented.kind === "checking" || lint.kind === "running" ? "pending" : "ready";
}

export function problemsOf(draft: TwinDraft): readonly TwinIssue[] {
  return choiceProblems(draft.doc);
}

export function pairsOf(draft: TwinDraft): readonly { readonly kind: KindFile; readonly choice: KindChoice }[] {
  return draft.members.flatMap(({ path, kind }) => {
    const choice = draft.doc.kinds.find((entry) => entry.kind === path);
    return choice === undefined ? [] : [{ kind, choice }];
  });
}

export function othersOf(draft: TwinDraft): readonly KindEntry[] {
  const taken = draft.members.map((member) => member.path);
  return draft.known.filter((entry) => !taken.includes(entry.path));
}

export function createBlock(draft: TwinDraft, device: DeviceState): string | null {
  if (isDirty(device)) return "Save the interface first: the twin is made from the saved device.";
  if (isBuilding(draft)) return "The twin is being created.";
  if (draft.save.kind === "write-failed" && draft.save.failure.kind === "conflict") {
    return "The choices file changed on disk. Reload it to continue.";
  }
  if (device.phase === "ready" && interfaceNeedsOf(device, draft).length > 0) {
    return "Set what the twin needs in the Interface step first.";
  }
  const [held] = draft.held.filter((entry) => !entry.adding);
  if (held !== undefined) return `${held.words}.`;
  return problemsOf(draft).length > 0 ? "Finish the choices marked in the form first." : null;
}

export function canCreate(draft: TwinDraft, device: DeviceState): boolean {
  return createBlock(draft, device) === null;
}
