import type {
  ConnectFailure,
  ControlFailure,
  DeviceFailure,
  Diagnostic,
  EditorFailure,
  LaunchFailure,
  PlantFailure,
  PolicyDefaultFailure,
  ReviewFailure,
  SessionFailure,
  TwinFailure,
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

type DeviceStep = "read" | "write" | "check" | "twin" | "capture";

const refusingSteps: readonly DeviceStep[] = ["check", "twin", "capture"];

export function deviceFailure(step: DeviceStep): FailureMapper<DeviceFailure> {
  return (error) => (error.kind === "http" ? deviceProblem(error.problem, step) : unreached(error));
}

function deviceProblem(problem: WireProblem, step: DeviceStep): DeviceFailure {
  const display = diagnostic(problem);
  if (isMissing(problem)) return { kind: "missing", diagnostic: display };
  if (problem.code === "conflict") return { kind: "conflict", diagnostic: display };
  if (refusingSteps.includes(step) && problem.status === 422) {
    return { kind: "refused", diagnostic: display, issues: problem.issues };
  }
  return serverFault(problem, display);
}

export function launchFailure(step: "start" | "status" | "stop" | "list"): FailureMapper<LaunchFailure> {
  return (error) => (error.kind === "http" ? launchProblem(error.problem, step) : unreached(error));
}

function launchProblem(problem: WireProblem, step: "start" | "status" | "stop" | "list"): LaunchFailure {
  const display = diagnostic(problem);
  return problem.status >= 500
    ? { kind: "unavailable", diagnostic: display }
    : { kind: "rejected", step, diagnostic: display };
}

export function connectFailure(error: RequestFailure): ConnectFailure {
  if (error.kind !== "http") return unreached(error);
  const display = diagnostic(error.problem);
  if (error.problem.code === "session_running") return { kind: "busy", diagnostic: display };
  if (error.problem.status === 422) return { kind: "refused", diagnostic: display, issues: error.problem.issues };
  return serverFault(error.problem, display);
}

export function policyDefaultFailure(error: RequestFailure): PolicyDefaultFailure {
  if (error.kind !== "http") return unreached(error);
  const display = diagnostic(error.problem);
  if (error.problem.code === "conflict") return { kind: "conflict", diagnostic: display };
  return serverFault(error.problem, display);
}

const controlKinds: ReadonlyMap<string, ControlFailure["kind"]> = new Map<string, ControlFailure["kind"]>([
  ["no_session", "no-session"],
  ["control_unavailable", "no-controls"],
  ["control_gone", "gone"],
  ["control_timeout", "timeout"],
  ["control_unreadable", "unreadable"],
]);

export function controlFailure(error: RequestFailure): ControlFailure {
  if (error.kind !== "http") return unreached(error);
  const { problem } = error;
  const fallback = problem.status >= 500 ? "unavailable" : "rejected";
  return { kind: controlKinds.get(problem.code) ?? fallback, diagnostic: diagnostic(problem) };
}

export function twinFailure(error: RequestFailure): TwinFailure {
  if (error.kind !== "http") return unreached(error);
  const kind = error.problem.status >= 500 ? "unavailable" : "refused";
  return { kind, diagnostic: diagnostic(error.problem) };
}

export function reviewFailure(error: RequestFailure): ReviewFailure {
  if (error.kind !== "http") return unreached(error);
  const display = diagnostic(error.problem);
  if (isMissing(error.problem)) return { kind: "missing", diagnostic: display };
  if (error.problem.code === "run_incomplete" || error.problem.code === "sealed")
    return { kind: "run-incomplete", diagnostic: display };
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
