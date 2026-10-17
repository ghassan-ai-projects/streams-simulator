import { failure as failed, success } from "../app/ports";
import type {
  Diagnostic,
  EditorFailure,
  LaunchFailure,
  PlantFailure,
  Outcome,
  ReviewFailure,
  SessionFailure,
  WorkspaceFailure,
} from "../app/ports";

export interface CannedIssue {
  readonly pointer: string;
  readonly message: string;
}

/** A protocol-shaped canned answer, the shape the bridge double drives the old generic surface with. */
export interface CannedFailure {
  readonly status: number;
  readonly code: string;
  readonly message: string;
  readonly issues: readonly CannedIssue[];
  readonly result?: unknown;
}

export type Canned<T> = Outcome<T, CannedFailure>;

export type PlantAction = "read" | "generate" | "write" | "compose";
export type EditorAction = "read" | "write" | "compose";
export type LaunchStep = "create" | "advance" | "inject" | "invoke" | "end";

export function failure(code: string, status = 401, message = "refused"): CannedFailure {
  return { status, code, message, issues: [] };
}

/** `unreachable` is the double's stand-in for a transport that never answered. */
function answered(canned: CannedFailure | null): CannedFailure | null {
  return canned === null || canned.code === "unreachable" ? null : canned;
}

function diagnostic(problem: CannedFailure | null): Diagnostic {
  return problem === null
    ? { message: "The bridge did not answer. Is streamsim canvas still running?", displayCode: "unreachable" }
    : { message: problem.message, displayCode: problem.code };
}

function missing(problem: CannedFailure): boolean {
  return problem.code === "not_found" || problem.code === "file_not_found";
}

export function settle<T, F>(outcome: Canned<T>, map: (problem: CannedFailure) => F): Outcome<T, F> {
  return outcome.ok ? success(outcome.value) : failed(map(outcome.failure));
}

/** A successful answer the double cannot project is a decode failure, exactly as the real adapter reports one. */
export function collect<T, F>(
  outcome: Canned<unknown>,
  map: (problem: CannedFailure) => F,
  read: (value: unknown) => T | null,
): Outcome<T, F> {
  if (!outcome.ok) return failed(map(outcome.failure));
  const value = read(outcome.value);
  return value === null ? failed(map(failure("unreadable", 200, "the bridge answered something else"))) : success(value);
}

export function sessionFailure(canned: CannedFailure | null): SessionFailure {
  const problem = answered(canned);
  if (problem === null) return { kind: "unreachable", diagnostic: diagnostic(null) };
  if (problem.code === "pairing_locked" || problem.code === "pairing_used") {
    return {
      kind: "pairing-closed",
      reason: problem.code === "pairing_locked" ? "locked" : "already-used",
      diagnostic: diagnostic(problem),
    };
  }
  return {
    kind: problem.code === "wrong_code" ? "code-rejected" : "unreachable",
    diagnostic: diagnostic(problem),
  };
}

export function workspaceFailure(canned: CannedFailure | null): WorkspaceFailure {
  const problem = answered(canned);
  if (problem === null) return { kind: "unavailable", diagnostic: diagnostic(null) };
  if (problem.code === "unpaired") return { kind: "unpaired", diagnostic: diagnostic(problem) };
  const display = diagnostic(problem);
  return problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "unreadable", diagnostic: display };
}

export function plantFailure(action: PlantAction, canned: CannedFailure | null): PlantFailure {
  const problem = answered(canned);
  if (problem === null) return { kind: "unavailable", diagnostic: diagnostic(null) };
  const display = diagnostic(problem);
  if (missing(problem)) return { kind: "missing", diagnostic: display };
  if (problem.code === "conflict") return { kind: "conflict", diagnostic: display };
  if (action === "generate" && problem.issues.length > 0) {
    return { kind: "invalid-composition", diagnostic: display, issues: problem.issues };
  }
  if (action === "compose" && problem.issues.length > 0) {
    return { kind: "compose-refused", diagnostic: display, issues: problem.issues };
  }
  return problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "unreadable", diagnostic: display };
}

export function editorFailure(action: EditorAction, canned: CannedFailure | null): EditorFailure {
  const problem = answered(canned);
  if (problem === null) return { kind: "unavailable", diagnostic: diagnostic(null) };
  const display = diagnostic(problem);
  if (missing(problem)) return { kind: "missing", diagnostic: display };
  if (problem.code === "conflict") return { kind: "conflict", diagnostic: display };
  if (action === "compose" && problem.issues.length > 0) {
    return { kind: "compose-refused", diagnostic: display, issues: problem.issues };
  }
  return problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "unreadable", diagnostic: display };
}

export function launchFailure(step: LaunchStep, canned: CannedFailure | null): LaunchFailure {
  const problem = answered(canned);
  if (problem === null) return { kind: "unavailable", diagnostic: diagnostic(null) };
  return { kind: "rejected", step, diagnostic: diagnostic(problem) };
}

export function reviewFailure(canned: CannedFailure | null): ReviewFailure {
  const problem = answered(canned);
  if (problem === null) return { kind: "unavailable", diagnostic: diagnostic(null) };
  if (missing(problem)) return { kind: "missing", diagnostic: diagnostic(problem) };
  if (problem.code === "run_incomplete") return { kind: "run-incomplete", diagnostic: diagnostic(problem) };
  const display = diagnostic(problem);
  return problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "unreadable", diagnostic: display };
}
