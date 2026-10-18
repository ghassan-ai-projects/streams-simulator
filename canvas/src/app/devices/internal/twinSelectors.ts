import { choiceProblems, serialiseChoices } from "../../../model/devices";
import type { KindChoice, KindFile, TwinIssue } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { Built, TwinDraft } from "./twinDraft";

export function isChoicesDirty(draft: TwinDraft): boolean {
  if (draft.doc === draft.saved) return false;
  return serialiseChoices(draft.doc) !== serialiseChoices(draft.saved);
}

export function isStale(draft: TwinDraft): boolean {
  return draft.build.kind === "built" && serialiseChoices(draft.doc) !== draft.build.fingerprint;
}

export function isBuilding(draft: TwinDraft): boolean {
  return draft.build.kind === "building";
}

export function builtOf(draft: TwinDraft): Built | null {
  return draft.build.kind === "built" && !isStale(draft) ? draft.build : null;
}

export function problemsOf(draft: TwinDraft): readonly TwinIssue[] {
  return choiceProblems(draft.doc);
}

export function pairsOf(draft: TwinDraft): readonly { readonly kind: KindFile; readonly choice: KindChoice }[] {
  return draft.members.flatMap(({ kind }) => {
    const choice = draft.doc.kinds.find((entry) => entry.kind === kind.kind);
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
  return problemsOf(draft).length > 0 ? "Finish the choices marked in the form first." : null;
}

export function canCreate(draft: TwinDraft, device: DeviceState): boolean {
  return createBlock(draft, device) === null;
}
