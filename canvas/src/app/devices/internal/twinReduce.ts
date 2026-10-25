import { resolveTwinIssues, serialiseKind } from "../../../model/devices";
import type { KindFile, Member, TwinChoices } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";
import type { Catalogue, Held, TwinDraft } from "./twinDraft";
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
    held: loaded.held,
    reviewed: [],
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

export function pendingOf(draft: TwinDraft): readonly Held[] {
  return draft.held.filter((held) => held.adding);
}

export function holdBack(
  draft: TwinDraft,
  waiting: Held,
  members: readonly Held[] = draft.held.filter((held) => !held.adding),
): TwinDraft {
  const others = pendingOf(draft).filter((held) => held.path !== waiting.path);
  const kept = members.filter((held) => held.path !== waiting.path);
  return { ...draft, held: [...kept, ...others, { ...waiting, adding: true }] };
}

interface Fresh {
  readonly suggested: Readonly<Record<string, unknown>> | null;
  readonly held: readonly Held[];
}

export function joinMembers(
  draft: TwinDraft,
  members: readonly Member[],
  instruments: Catalogue,
  next: Fresh,
): TwinDraft {
  const grown = withMembers({ ...draft, suggested: next.suggested ?? draft.suggested }, members);
  const pending = pendingOf(draft).filter((held) => !members.some((member) => member.path === held.path));
  return { ...grown, instruments, held: [...next.held, ...pending] };
}
