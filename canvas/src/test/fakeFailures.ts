import { failure as failed, success } from "../app/ports";
import * as real from "../adapters/bridge";
import type { RequestFailure } from "../adapters/bridge";
import type {
  EditorFailure,
  LaunchFailure,
  PlantFailure,
  Outcome,
  ReviewFailure,
  SessionFailure,
  WorkspaceFailure,
} from "../app/ports";

interface CannedIssue {
  readonly pointer: string;
  readonly message: string;
}

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

export function cannedFailure(code: string, status = 401, message = "refused"): CannedFailure {
  return { status, code, message, issues: [] };
}

function asRequest(canned: CannedFailure | null): RequestFailure {
  return canned === null || canned.code === "unreachable"
    ? { kind: "transport" }
    : { kind: "http", problem: { ...canned, result: canned.result ?? null } };
}

export function settle<T, F>(outcome: Canned<T>, map: (problem: CannedFailure) => F): Outcome<T, F> {
  return outcome.ok ? success(outcome.value) : failed(map(outcome.failure));
}

export function collect<T, F>(
  outcome: Canned<unknown>,
  map: (problem: CannedFailure) => F,
  read: (value: unknown) => T | null,
): Outcome<T, F> {
  if (!outcome.ok) return failed(map(outcome.failure));
  const value = read(outcome.value);
  return value === null
    ? failed(map(cannedFailure("unreadable", 200, "the bridge answered something else")))
    : success(value);
}

export function sessionFailure(canned: CannedFailure | null): SessionFailure {
  return real.sessionFailure(asRequest(canned));
}

export function workspaceFailure(canned: CannedFailure | null): WorkspaceFailure {
  return real.workspaceFailure(asRequest(canned));
}

export function plantFailure(action: PlantAction, canned: CannedFailure | null): PlantFailure {
  return real.plantFailure(action)(asRequest(canned));
}

export function editorFailure(action: EditorAction, canned: CannedFailure | null): EditorFailure {
  return real.editorFailure(action)(asRequest(canned));
}

export function launchFailure(step: LaunchStep, canned: CannedFailure | null): LaunchFailure {
  return real.launchFailure(step)(asRequest(canned));
}

export function reviewFailure(canned: CannedFailure | null): ReviewFailure {
  return real.reviewFailure(asRequest(canned));
}
