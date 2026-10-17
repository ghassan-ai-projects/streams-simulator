import type {
  DeviceFailure,
  Diagnostic,
  EditorFailure,
  LaunchFailure,
  PlantFailure,
  ReviewFailure,
  SessionFailure,
  WorkspaceFailure,
} from "../../../app/ports";
import type { FailureMapper, RequestFailure, WireProblem } from "./http";

type Unreached = Exclude<RequestFailure, { readonly kind: "http" }>;

export function sessionFailure(error: RequestFailure): SessionFailure {
  if (error.kind !== "http") return { kind: "unreachable", diagnostic: diagnostic(null) };
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
  if (error.kind !== "http") return unreached(error);
  const display = diagnostic(error.problem);
  if (error.problem.code === "unpaired") return { kind: "unpaired", diagnostic: display };
  return serverFault(error.problem, display);
}

export function plantFailure(action: "read" | "generate" | "write" | "compose"): FailureMapper<PlantFailure> {
  return (error) => (error.kind === "http" ? plantProblem(error.problem, action) : unreached(error));
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
  return serverFault(problem, display);
}

export function editorFailure(action: "read" | "write" | "compose"): FailureMapper<EditorFailure> {
  return plantFailure(action);
}

export function deviceFailure(step: "read" | "write" | "check"): FailureMapper<DeviceFailure> {
  return (error) => (error.kind === "http" ? deviceProblem(error.problem, step) : unreached(error));
}

function deviceProblem(problem: WireProblem, step: "read" | "write" | "check"): DeviceFailure {
  const display = diagnostic(problem);
  if (isMissing(problem)) return { kind: "missing", diagnostic: display };
  if (problem.code === "conflict") return { kind: "conflict", diagnostic: display };
  if (step === "check" && problem.status === 422) {
    return { kind: "refused", diagnostic: display, issues: problem.issues };
  }
  return serverFault(problem, display);
}

export function launchFailure(step: "create" | "advance" | "inject" | "invoke" | "end"): FailureMapper<LaunchFailure> {
  return (error) =>
    error.kind === "http" ? { kind: "rejected", step, diagnostic: diagnostic(error.problem) } : unreached(error);
}

export function reviewFailure(error: RequestFailure): ReviewFailure {
  if (error.kind !== "http") return unreached(error);
  const display = diagnostic(error.problem);
  if (isMissing(error.problem)) return { kind: "missing", diagnostic: display };
  if (error.problem.code === "run_incomplete") return { kind: "run-incomplete", diagnostic: display };
  return serverFault(error.problem, display);
}

function unreached(error: Unreached) {
  return { kind: error.kind === "transport" ? "unavailable" : "unreadable", diagnostic: diagnostic(null) } as const;
}

function serverFault(problem: WireProblem, display: Diagnostic) {
  return { kind: problem.status >= 500 ? "unavailable" : "unreadable", diagnostic: display } as const;
}

function diagnostic(problem: WireProblem | null): Diagnostic {
  return problem === null
    ? { message: "The bridge did not answer. Is streamsim canvas still running?", displayCode: "unreachable" }
    : { message: problem.message, displayCode: problem.code };
}

function isMissing(problem: WireProblem): boolean {
  return problem.code === "not_found" || problem.code === "file_not_found";
}
