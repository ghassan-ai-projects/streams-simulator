import { issueText, resolveIssues } from "./issues";
import { neededPhrase } from "./issueWords";
import type { FieldIssue, KindFile, RawIssue, Selection } from "./types";

export interface HeldIssue {
  readonly owner: "reading" | "command" | "device";
  readonly ownerName: string;
  readonly setting: string;
  readonly sentence: string;
  readonly original: string;
  readonly part: FieldIssue["part"];
  readonly key: string;
  readonly field: string;
}

const shownReasons = 3;

function capital(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

function ownerKind(part: FieldIssue["part"]): HeldIssue["owner"] {
  if (part === "metric") return "reading";
  return part === "command" ? "command" : "device";
}

function nameOf(kind: KindFile, placed: FieldIssue): string {
  const entries = placed.part === "metric" ? kind.metrics : placed.part === "command" ? kind.commands : [];
  return entries.find((entry) => entry.key === placed.key)?.name ?? "";
}

function phrase(owner: HeldIssue["owner"], name: string): string {
  if (owner === "device") return "the device";
  return name === "" ? `an unnamed ${owner}` : `${owner} '${name}'`;
}

function sentenceOf(raw: RawIssue, owner: string): string {
  const keyword = raw.keyword ?? "";
  const member = raw.member ?? "";
  if (keyword === "twin_member_unsupported") {
    const setting = member === "" ? "A setting" : `The setting ${member}`;
    return `${setting} of ${owner} is not reproduced by a twin.`;
  }
  const needed = neededPhrase(keyword, member);
  if (needed !== null) return `${capital(owner)} needs ${needed}.`;
  return `${capital(owner)}: ${issueText(raw)}.`;
}

function heldIssue(kind: KindFile, raw: RawIssue, placed: FieldIssue): HeldIssue {
  const owner = ownerKind(placed.part);
  const ownerName = nameOf(kind, placed);
  const sentence = sentenceOf(raw, phrase(owner, ownerName));
  const { part, key, field } = placed;
  return { owner, ownerName, setting: raw.member ?? "", sentence, original: raw.message, part, key, field };
}

export function heldIssues(kind: KindFile, issues: readonly RawIssue[]): readonly HeldIssue[] {
  const placed = resolveIssues(kind, issues);
  return issues.flatMap((raw, index) => {
    const at = placed[index];
    return at === undefined ? [] : [heldIssue(kind, raw, at)];
  });
}

export function heldReason(issues: readonly HeldIssue[]): string {
  const lines = issues.slice(0, shownReasons).map(({ sentence }) => sentence.replace(/\.$/, ""));
  const more = issues.length - shownReasons;
  const reasons = more > 0 ? [...lines, `and ${String(more)} more`] : lines;
  return reasons.map((reason) => reason.charAt(0).toLowerCase() + reason.slice(1)).join("; ");
}

export function leftOutWords(device: string): string {
  return `${device} is left out of the twin as a whole, so none of its readings or commands are reproduced. Fix the device, then select Retry to include it.`;
}

export function blockingWords(device: string): string {
  return `${device} is part of this twin but cannot be reproduced yet, so the twin cannot be created until it is fixed or left out.`;
}

export function heldSelection(issue: HeldIssue): Selection {
  if (issue.part === "metric" || issue.part === "command") return { part: issue.part, key: issue.key };
  return { part: "kind" };
}
