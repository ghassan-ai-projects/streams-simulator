import { heldReason, partialWords, stemOf } from "../../../model/devices";
import type { Held, TwinDraft } from "./twinDraft";

export interface Exclusion {
  readonly name: string;
  readonly reason: string;
}

export interface Scope {
  readonly included: readonly string[];
  readonly excluded: readonly Exclusion[];
  readonly blocked: readonly Exclusion[];
}

function reasonOf(held: Held): string {
  if (held.issues.length > 0) return heldReason(held.issues);
  return held.words.startsWith(held.name) ? held.words.slice(held.name.length).trim() : held.words;
}

function exclusion(held: Held): Exclusion {
  return { name: held.name, reason: reasonOf(held) };
}

export function scopeOf(draft: TwinDraft): Scope {
  return {
    included: draft.members.map((member) => member.kind.kind || stemOf(member.path)),
    excluded: draft.held.filter((held) => held.adding).map(exclusion),
    blocked: draft.held.filter((held) => !held.adding).map(exclusion),
  };
}

export function partialOf(draft: TwinDraft): string {
  return partialWords(scopeOf(draft).excluded);
}
