import { counted, shortDigest } from "../../shared/words";
import type { PlanIssue, PolicySource, SessionDraft } from "./session";
import { firstSessionDraft, policyLines } from "./session";

export function isPolicyIssue(issue: PlanIssue): boolean {
  return onPolicy.test(issue.pointer);
}

export const refusedFieldWords = "streamsim refused the plan: fix the marked field and check it again.";

export interface PolicyNote {
  readonly digest: number | null;
  readonly issue: PlanIssue;
}

const onPolicy = /^\/policies(?:\/(\d+))?$/;

export function policyNotes(issues: readonly PlanIssue[]): readonly PolicyNote[] {
  return issues.flatMap((issue) => {
    const place = onPolicy.exec(issue.pointer);
    if (place === null) return [];
    const index = place[1];
    return [{ digest: index === undefined ? null : Number(index) + 1, issue }];
  });
}

const digestWords = "A policy digest is sha256: and 64 lower-case hex digits.";
const tooManyWords = "List at most 8 policy digests.";

const keywordWords: Readonly<Record<string, string>> = {
  pattern: digestWords,
  policy_malformed: digestWords,
  maxItems: tooManyWords,
  too_many_policies: tooManyWords,
  policy_listed_twice: "This digest is already listed.",
  policy_reserved: "This digest is reserved for the node's own stops.",
};

export function policyNoteWords({ digest, issue }: PolicyNote): string {
  const words = keywordWords[issue.keyword ?? ""] ?? issue.message;
  return digest === null ? words : `Digest ${String(digest)}: ${words}`;
}

export const policyHint =
  "One per line, at most 8. A command to a simulated device runs only when the host's policy digest is listed; " +
  "safe stops always run. Real boards are checked by the gateway's own list.";

const noPoliciesWords = "Allowed policies: none. Commands to simulated devices are refused, except safe stops.";

export function policyListWords(policies: readonly string[]): string {
  return policies.length === 0 ? noPoliciesWords : `Allowed policies: ${String(policies.length)}`;
}

export function recordedPolicyWords(policies: readonly string[] | null): string {
  return policies === null ? "Allowed policies: not recorded." : policyListWords(policies);
}

export function withPolicyText(draft: SessionDraft, policies: string): SessionDraft {
  return { ...draft, policies, policiesFrom: "operator" };
}

export function withPolicyListed(draft: SessionDraft, digest: string): SessionDraft {
  const lines = policyLines(draft.policies);
  if (lines.includes(digest)) return draft;
  return withPolicyText(draft, [...lines, digest].join("\n"));
}

export function filledFromDefault(draft: SessionDraft, policies: readonly string[]): SessionDraft {
  return { ...draft, policies: policies.join("\n"), policiesFrom: "default" };
}

export const prefilledWords = "Pre-filled from the workspace default.";

export const unreadableDefaultWords = "The workspace default could not be read.";

export const saveDefaultHint =
  "Saves the list streamsim accepted, sorted, as the default for new session forms. An empty list clears it. " +
  "A running session never changes.";

type DefaultSave = "saving" | "saved" | "conflict" | "failed";

export function saveDefaultWords(save: DefaultSave, problem: string, saved: readonly string[]): string {
  if (save === "saving") return "Saving the workspace default…";
  if (save === "conflict")
    return "The workspace default changed elsewhere, so nothing was saved. Check the plan again.";
  if (save === "failed") return `The workspace default was not saved: ${problem}`;
  return saved.length === 0
    ? "The workspace default is now empty."
    : `Saved ${counted(saved.length, "policy digest")} as the workspace default.`;
}

export function keptWords(policies: string): string {
  return `Kept from the previous run: ${counted(policyLines(policies).length, "digest")}.`;
}

export function keepForNextRun(draft: SessionDraft): SessionDraft {
  const kept = policyLines(draft.policies).length > 0;
  return { ...firstSessionDraft, policies: draft.policies, policiesFrom: kept ? "previous" : "operator" };
}

export function keptFromRecord(policies: readonly string[]): SessionDraft {
  return keepForNextRun({ ...firstSessionDraft, policies: policies.join("\n") });
}

const trustOrigins: Readonly<Record<PolicySource, string>> = {
  default: "from the workspace default",
  operator: "you entered",
  previous: "kept from the previous run",
};

const trustShown = 3;

export function trustWords(draft: SessionDraft): string {
  const lines = policyLines(draft.policies);
  if (lines.length === 0) return "";
  const shown = lines.slice(0, trustShown).map(shortDigest);
  if (lines.length > trustShown) shown.push(`and ${String(lines.length - trustShown)} more`);
  return `Trusting ${counted(lines.length, "policy digest")} ${trustOrigins[draft.policiesFrom]}: ${shown.join(", ")}`;
}
