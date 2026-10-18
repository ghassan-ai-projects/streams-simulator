import { issueText } from "./issues";
import type { RawIssue } from "./types";

type Environment = "hardware" | "simulated";

type Comparison =
  { readonly status: "new" } | { readonly status: "same" | "differs" | "unreadable"; readonly file: string };

interface Scan {
  readonly birthsSeen: number;
  readonly taken: { readonly line: number };
  readonly laterBirths: readonly { readonly sameContent: boolean }[];
  readonly issues: readonly RawIssue[];
}

interface RefusalOutline {
  readonly heading: string;
  readonly sentence: string;
  readonly next: string;
}

const maxCaptureBytes = 1024 * 1024;

export const claimWords = "A board's announcement is a claim. Nothing is saved until you review and accept it.";

const tooLarge =
  "That is larger than 1 MiB. Paste only the board's announcement, or choose a recorded run to read it from.";

const copyAgain = "Copy the announcement again from the board.";

const refusedHeading = "The board's announcement was not accepted";

const outlines: Readonly<Record<string, RefusalOutline>> = {
  no_birth: {
    heading: "No announcement found",
    sentence: "There is no announcement from a board in what was read.",
    next: "Copy the announcement again from the board, or choose a run that was recorded with one.",
  },
  json: {
    heading: "That is not valid JSON",
    sentence: "What was pasted cannot be read as JSON.",
    next: "Copy the whole announcement again from the board.",
  },
  kind_digest_mismatch: {
    heading: refusedHeading,
    sentence: "The announcement does not match its own fingerprint, so it may have been changed or damaged.",
    next: copyAgain,
  },
  sealed_stream: {
    heading: refusedHeading,
    sentence: "This record belongs to the sealed part of a run, which is never read here.",
    next: "Choose the run itself, or paste the announcement the board sent.",
  },
  stream_not_observer: {
    heading: refusedHeading,
    sentence: "This record did not come from the observer's side of a run.",
    next: "Choose the run itself, or paste the announcement the board sent.",
  },
};

const refusedOutline: RefusalOutline = {
  heading: refusedHeading,
  sentence: "The announcement cannot be trusted as it stands.",
  next: copyAgain,
};

const failureOutlines: Readonly<Record<string, RefusalOutline>> = {
  missing: {
    heading: "No announcement found",
    sentence: "This run has no recorded announcement.",
    next: "Choose another run, or paste the announcement the board sent.",
  },
  unavailable: {
    heading: "The bridge is unavailable",
    sentence: "The bridge did not answer. Is streamsim canvas still running?",
    next: "Start it again, then read the announcement once more.",
  },
  unreadable: {
    heading: "The announcement could not be read",
    sentence: "The bridge answered with something this page does not understand.",
    next: "Try again; if it keeps failing, copy a shorter announcement.",
  },
};

export function originWords(environment: Environment): string {
  return environment === "hardware" ? "Real board" : "Simulated";
}

export function birthWords(bdSeq: number): string {
  return bdSeq === 0 ? "" : `Board restart count: ${String(bdSeq)}`;
}

export function sourceWords(source: { readonly content: string } | { readonly run: string }): string {
  return "run" in source ? `the run ${source.run}` : "the pasted text";
}

function fileName(path: string): string {
  return path.split("/").at(-1) ?? path;
}

export function comparisonWords(comparison: Comparison): string {
  if (comparison.status === "new") return "New device";
  const file = fileName(comparison.file);
  if (comparison.status === "unreadable") return "Cannot compare";
  return comparison.status === "same" ? `Same as ${file}` : `Differs from ${file}`;
}

export function collisionWords(file: string): string {
  return `Looks like ${fileName(file)}, which differs only by capital letters.`;
}

export function unreadableWords(file: string): string {
  return `The saved file ${fileName(file)} cannot be read, so it cannot be compared. Repair or remove it, then read the announcement again.`;
}

export function captureSizeProblem(bytes: number): string {
  return bytes > maxCaptureBytes ? tooLarge : "";
}

export function captureTextProblem(source: string): string {
  if (source.length * 3 <= maxCaptureBytes) return "";
  return captureSizeProblem(new TextEncoder().encode(source).length);
}

export function refusalOutline(kind: string, keywords: readonly string[]): RefusalOutline {
  if (kind !== "refused") return failureOutlines[kind] ?? refusedOutline;
  const known = keywords.find((keyword) => Object.hasOwn(outlines, keyword));
  return (known === undefined ? undefined : outlines[known]) ?? refusedOutline;
}

export function refusalLines(
  issues: readonly RawIssue[],
): readonly { readonly text: string; readonly message: string }[] {
  return issues.map((issue) => ({ text: issueText(issue), message: issue.message }));
}

const scanProblems: Readonly<Record<string, string>> = {
  birth_line_unreadable: "A line that names an announcement could not be read.",
  capability_digest_mismatch:
    "The board's capability fingerprint does not match its devices; treat this announcement with care.",
  scan_truncated: "Only the start of this log was read.",
};

function laterWords({ birthsSeen, taken, laterBirths }: Scan, truncated: boolean): string[] {
  if (birthsSeen < 2) return [];
  const count = `${truncated ? "at least " : ""}${String(birthsSeen)}`;
  const differ = laterBirths.some((later) => !later.sameContent);
  const rest = differ ? "Later ones differ." : "Later ones describe the same devices.";
  return [`This log holds ${count} announcements; showing the first (line ${String(taken.line)}). ${rest}`];
}

export function scanNotices(scan: Scan): readonly string[] {
  const keywords = scan.issues.map((issue) => issue.keyword ?? "");
  const truncated = keywords.includes("scan_truncated");
  const problems = keywords.filter((keyword) => Object.hasOwn(scanProblems, keyword));
  const unique = [...new Set(problems)].map((keyword) => scanProblems[keyword] ?? "");
  const later = laterWords(scan, truncated);
  const rebirth = keywords.includes("rebirth_differs") && later.length === 0;
  return [...later, ...(rebirth ? ["A later announcement describes other devices than the first."] : []), ...unique];
}

function trimmed(claim: string): string {
  const parts = claim.split("/").filter((part) => part !== "");
  if (parts.includes("args_schema")) return "argument limits";
  return (parts.at(-1) ?? "").replaceAll("_", " ");
}

export function claimedWords(claims: readonly string[]): string {
  const named = [...new Set(claims.map(trimmed).filter((claim) => claim !== ""))];
  if (named.length === 0) return "";
  return `Taken from the board's announcement, not verified: ${named.join(", ")}.`;
}
