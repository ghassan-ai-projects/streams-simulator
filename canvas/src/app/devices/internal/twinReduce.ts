import { resolveTwinIssues, serialiseKind } from "../../../model/devices";
import type { KindFile, Member, TwinChoices } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";
import type { TwinDraft } from "./twinDraft";
import type { Loaded } from "./twinLoad";
import { reconciled } from "./twinLoad";

export function fresh(path: string, loaded: Loaded): TwinDraft {
  return {
    phase: "ready",
    path,
    choicesPath: loaded.choicesPath,
    members: loaded.members,
    known: loaded.known,
    instruments: loaded.instruments,
    suggested: loaded.suggested,
    onDisk: loaded.onDisk,
    touched: loaded.touched,
    edited: false,
    saved: loaded.saved,
    written: loaded.written,
    doc: loaded.doc,
    save: { kind: "idle" },
    build: { kind: "idle" },
    issues: [],
  };
}

export function withMembers(draft: TwinDraft, members: readonly Member[]): TwinDraft {
  const doc = reconciled(draft, members);
  return { ...draft, members, doc, touched: true, edited: true, issues: [] };
}

export function afterRefusal(draft: TwinDraft, failure: DeviceFailure, sent: TwinChoices): TwinDraft {
  if (failure.kind === "conflict") return { ...draft, build: { kind: "exists" } };
  if (failure.kind !== "refused") return { ...draft, build: { kind: "failed", failure } };
  const issues = resolveTwinIssues(sent, failure.issues, draft.members);
  return { ...draft, build: { kind: "refused", failure }, issues };
}

export function sameKind(a: KindFile, b: KindFile): boolean {
  return a === b || serialiseKind(a) === serialiseKind(b);
}

export const missingLead: DeviceFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "kind_unreadable", message: "The device is not open." },
};

export function afterFlush(draft: TwinDraft, saved: TwinChoices, written: string): TwinDraft {
  const edited = draft.doc === saved ? false : draft.edited;
  return { ...draft, saved, written, save: { kind: "idle" }, edited };
}
