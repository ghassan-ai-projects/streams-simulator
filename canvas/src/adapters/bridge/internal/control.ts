import type { ControlFailure, Outcome } from "../../../app/ports";
import type { ControlReply, ControlRequest, ControlScope } from "../../../model/review";
import { controlFailure } from "./failures";
import type { Transport } from "./http";
import { action } from "./http";
import { controlReply } from "./interruptDecoder";

function wireScope(scope: ControlScope): Readonly<Record<string, string>> {
  return scope.kind === "all" ? { kind: "all" } : { kind: scope.kind, id: scope.id };
}

function controlArgs(request: ControlRequest): Readonly<Record<string, unknown>> {
  switch (request.op) {
    case "hold":
    case "release":
    case "estop":
      return { op: request.op, scope: wireScope(request.scope) };
    case "approve_mode":
      return { op: request.op, scope: wireScope(request.scope), enabled: request.enabled };
    case "approve":
    case "deny":
      return { op: request.op, command_id: request.commandId, content_digest: request.contentDigest };
    case "state":
      return { op: request.op };
  }
}

export function controlSession(
  transport: Transport,
  request: ControlRequest,
): Promise<Outcome<ControlReply, ControlFailure>> {
  return action(transport, "session.control", controlArgs(request), controlReply, controlFailure);
}
