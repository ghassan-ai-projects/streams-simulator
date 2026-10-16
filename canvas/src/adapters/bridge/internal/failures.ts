import type {
  Diagnostic,
  EditorFailure,
  LaunchFailure,
  PlantFailure,
  ReviewFailure,
  SessionFailure,
  WorkspaceFailure,
} from "../../../app/ports/index";
import type { FailureMapper, RequestFailure, WireProblem } from "./http";

export function sessionFailure(error: RequestFailure): SessionFailure {
  if (error.kind === "transport") return { kind: "unreachable", diagnostic: diagnostic(null) };
  if (error.kind === "unreadable") return { kind: "unreachable", diagnostic: diagnostic(null) };
  const detail = error.problem;
  if (detail.code === "pairing_locked" || detail.code === "pairing_used") {
    return {
      kind: "pairing-closed",
      reason: detail.code === "pairing_locked" ? "locked" : "already-used",
      diagnostic: diagnostic(detail),
    };
  }
  return {
    kind: detail.code === "wrong_code" ? "code-rejected" : "unreachable",
    diagnostic: diagnostic(detail),
  };
}

export function workspaceFailure(error: RequestFailure): WorkspaceFailure {
  if (error.kind === "transport") return { kind: "unavailable", diagnostic: diagnostic(null) };
  if (error.kind === "unreadable") return { kind: "unreadable", diagnostic: diagnostic(null) };
  if (error.problem.code === "unpaired") return { kind: "unpaired", diagnostic: diagnostic(error.problem) };
  const display = diagnostic(error.problem);
  return error.problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "unreadable", diagnostic: display };
}

export function plantFailure(action: "read" | "generate" | "write" | "compose"): FailureMapper<PlantFailure> {
  return (error) => {
    if (error.kind === "transport") return { kind: "unavailable", diagnostic: diagnostic(null) };
    if (error.kind === "unreadable") return { kind: "unreadable", diagnostic: diagnostic(null) };
    return plantProblem(error.problem, action);
  };
}

function plantProblem(problem: WireProblem, action: "read" | "generate" | "write" | "compose"): PlantFailure {
  const display = diagnostic(problem);
  if (isMissing(problem)) return { kind: "missing", diagnostic: display };
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

export function editorFailure(action: "read" | "write" | "compose"): FailureMapper<EditorFailure> {
  return (error) => {
    if (error.kind === "transport") return { kind: "unavailable", diagnostic: diagnostic(null) };
    if (error.kind === "unreadable") return { kind: "unreadable", diagnostic: diagnostic(null) };
    const problem = error.problem;
    const display = diagnostic(problem);
    if (isMissing(problem)) return { kind: "missing", diagnostic: display };
    if (problem.code === "conflict") return { kind: "conflict", diagnostic: display };
    if (action === "compose" && problem.issues.length > 0) {
      return { kind: "compose-refused", diagnostic: display, issues: problem.issues };
    }
    return problem.status >= 500
      ? { kind: "unavailable", diagnostic: display }
      : { kind: "unreadable", diagnostic: display };
  };
}

export function launchFailure(step: "create" | "advance" | "inject" | "invoke" | "end"): FailureMapper<LaunchFailure> {
  return (error) => {
    if (error.kind === "transport") return { kind: "unavailable", diagnostic: diagnostic(null) };
    if (error.kind === "unreadable") return { kind: "unreadable", diagnostic: diagnostic(null) };
    return { kind: "rejected", step, diagnostic: diagnostic(error.problem) };
  };
}

export function reviewFailure(error: RequestFailure): ReviewFailure {
  if (error.kind === "transport") return { kind: "unavailable", diagnostic: diagnostic(null) };
  if (error.kind === "unreadable") return { kind: "unreadable", diagnostic: diagnostic(null) };
  if (error.problem.code === "not_found" || error.problem.code === "file_not_found") {
    return { kind: "missing", diagnostic: diagnostic(error.problem) };
  }
  if (error.problem.code === "run_incomplete") {
    return { kind: "run-incomplete", diagnostic: diagnostic(error.problem) };
  }
  const display = diagnostic(error.problem);
  return error.problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "unreadable", diagnostic: display };
}

function diagnostic(problem: WireProblem | null): Diagnostic {
  return problem === null
    ? { message: "The bridge did not answer. Is streamsim canvas still running?", displayCode: "unreachable" }
    : { message: problem.message, displayCode: problem.code };
}

function isMissing(problem: WireProblem): boolean {
  return problem.code === "not_found" || problem.code === "file_not_found";
}
