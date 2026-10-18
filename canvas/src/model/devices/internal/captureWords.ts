import { issueText } from "./issues";
import { devicePath } from "./rules";
import type { RawIssue } from "./types";

type Environment = "hardware" | "simulated";

type Comparison = { readonly status: "new" } | { readonly status: "same" | "differs"; readonly file: string };

interface RefusalLine {
  readonly text: string;
  readonly message: string;
}

const maxCaptureBytes = 1024 * 1024;

export const claimWords = "A board's announcement is a claim. Nothing is saved until you review and accept it.";

const tooLarge =
  "That is larger than 1 MiB. Paste only the board's announcement, or choose a recorded run to read it from.";

export function originWords(environment: Environment): string {
  return environment === "hardware" ? "Real board" : "Simulated";
}

export function birthWords(bdSeq: number): string {
  return `Announcement number ${String(bdSeq)}`;
}

function fileName(path: string): string {
  return path.split("/").at(-1) ?? path;
}

export function comparisonWords(comparison: Comparison): string {
  if (comparison.status === "new") return "New device";
  const file = fileName(comparison.file);
  return comparison.status === "same" ? `Same as ${file}` : `Differs from ${file}`;
}

export function captureSizeProblem(bytes: number): string {
  return bytes > maxCaptureBytes ? tooLarge : "";
}

export function captureTextProblem(source: string): string {
  return captureSizeProblem(new TextEncoder().encode(source).length);
}

export function fileSafeName(name: string): boolean {
  return name !== "" && !/[\\/]/.test(name) && !name.startsWith(".");
}

export function capturePath(kinds: readonly { readonly path: string; readonly kind: string }[], name: string): string {
  return kinds.find((entry) => entry.kind === name)?.path ?? devicePath(name);
}

export function refusalLines(issues: readonly RawIssue[]): readonly RefusalLine[] {
  return issues.map((issue) => ({ text: issueText(issue), message: issue.message }));
}
