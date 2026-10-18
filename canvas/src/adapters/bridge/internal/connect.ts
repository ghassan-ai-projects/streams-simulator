import type { ConnectGateway, ConnectStatus, PlannedInstance, SessionCheck } from "../../../app/ports";
import { isRecord, list, record, text } from "../../../model/shared/json";
import { connectFailure } from "./failures";
import { followRun } from "./follow";
import type { Transport } from "./http";
import { action, issueList } from "./http";

const roles = ["real", "simulated"] as const;

function roleOf(value: unknown): PlannedInstance["role"] | null {
  return roles.find((role) => role === value) ?? null;
}

function planned(value: unknown): PlannedInstance | null {
  const entry = record(value);
  const role = roleOf(entry.origin);
  if (role === null) return null;
  const { instance, kind, board, device } = entry;
  return { instance: text(instance), role, kind: text(kind), board: text(board), device: text(device) };
}

function sessionCheck(value: unknown): SessionCheck | null {
  if (!isRecord(value) || !Array.isArray(value.instances)) return null;
  const instances = list(value.instances).map(planned);
  if (instances.some((instance) => instance === null)) return null;
  return {
    domain: text(value.domain),
    mixed: value.environment === "mixed",
    instances: instances.filter((instance): instance is PlannedInstance => instance !== null),
  };
}

interface Facts {
  readonly run: string;
  readonly domain: string;
  readonly check: SessionCheck | null;
}

function failedStatus(facts: Facts, failure: Record<string, unknown>): ConnectStatus {
  const diagnostic = {
    message: text(failure.message) || "The session ended with an error and left no message.",
    displayCode: text(failure.code) || "session_failed",
  };
  return { state: "failed", ...facts, diagnostic, issues: issueList(failure.issues) };
}

const lastingStates = ["starting", "running", "stopping", "ended"] as const;

function lastingState(value: unknown): (typeof lastingStates)[number] | null {
  return lastingStates.find((state) => state === value) ?? null;
}

function connectStatus(value: unknown): ConnectStatus | null {
  if (!isRecord(value)) return null;
  if (value.state === "none") return { state: "none" };
  const facts = { run: text(value.run), domain: text(value.domain), check: sessionCheck(value.plan) };
  const state = lastingState(value.state);
  if (state !== null) return { state, ...facts };
  return value.state === "failed" ? failedStatus(facts, record(value.failure)) : null;
}

export function connectGateway(transport: Transport): ConnectGateway {
  return {
    checkSession: (plan) => action(transport, "session.check", { content: plan }, sessionCheck, connectFailure),
    startSession: (plan) => action(transport, "session.start", { content: plan }, connectStatus, connectFailure),
    sessionStatus: () => action(transport, "session.status", {}, connectStatus, connectFailure),
    stopSession: () => action(transport, "session.stop", {}, connectStatus, connectFailure),
    followSession: (run, follower) => followRun(transport, run, follower),
  };
}
