import type { ConnectGateway, ConnectStatus, PlannedInstance, SessionCheck } from "../../../app/ports";
import { isRecord, list, record, text } from "../../../model/shared/json";
import { connectFailure } from "./failures";
import { followRun } from "./follow";
import type { Transport } from "./http";
import { action, issueList } from "./http";

function planned(value: unknown): PlannedInstance {
  const entry = record(value);
  return {
    instance: text(entry.instance),
    role: entry.origin === "real" ? "real" : "simulated",
    kind: text(entry.kind),
    board: text(entry.board),
    device: text(entry.device),
  };
}

function sessionCheck(value: unknown): SessionCheck | null {
  if (!isRecord(value) || !Array.isArray(value.instances)) return null;
  return {
    domain: text(value.domain),
    mixed: value.environment === "mixed",
    instances: list(value.instances).map(planned),
  };
}

function failedStatus(run: string, failure: Record<string, unknown>): ConnectStatus {
  const diagnostic = {
    message: text(failure.message) || "The session ended with an error and left no message.",
    displayCode: text(failure.code) || "session_failed",
  };
  return { state: "failed", run, diagnostic, issues: issueList(failure.issues) };
}

function connectStatus(value: unknown): ConnectStatus | null {
  if (!isRecord(value)) return null;
  const run = text(value.run);
  if (value.state === "none") return { state: "none" };
  if (value.state === "running" || value.state === "ended") return { state: value.state, run };
  return value.state === "failed" ? failedStatus(run, record(value.failure)) : null;
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
