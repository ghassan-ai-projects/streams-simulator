import { durationOf } from "../../shared/duration";
import { counted } from "../../shared/words";
import type {
  ControlFailureKind,
  ControlOp,
  ControlReply,
  ControlRequest,
  ControlScope,
  InterruptMode,
  PendingApproval,
} from "./interrupts";
import { refusalWords } from "./sessionWords";

interface OutcomeLine {
  readonly tone: "ok" | "refused" | "failed";
  readonly text: string;
}

export type StateCheck = "none" | "pending" | "read" | "failed";

export interface NodeNow {
  readonly worldNs: number;
  readonly wallMs: number;
}

const modeWords: Readonly<Record<InterruptMode, string>> = {
  hold: "Held",
  approve: "Approving each command",
};

const opWords: Readonly<Record<ControlOp, string>> = {
  hold: "Hold",
  release: "Release",
  approve_mode: "Approve each command",
  approve: "Approve",
  deny: "Deny",
  estop: "Emergency stop",
  state: "Read the state",
};

const failureWords: Readonly<Record<ControlFailureKind, string>> = {
  "no-session": "No connected session is running, so nothing was sent.",
  "no-controls": "This session has no control socket, so the node cannot be controlled.",
  gone: "The node's control socket is not there. Nothing was confirmed.",
  timeout: "The node did not answer in time. The control may still apply afterwards.",
  unavailable: "The bridge did not answer. Nothing was confirmed.",
  unreadable: "The answer could not be read. Nothing was confirmed.",
};

interface ModeFacts {
  readonly mode: InterruptMode;
  readonly latched: boolean;
}

export function modeLine({ mode, latched }: ModeFacts): string {
  if (mode === "approve") return modeWords.approve;
  return latched ? "Held · latched by emergency stop" : modeWords.hold;
}

function scopeWords(scope: ControlScope): string {
  if (scope.kind === "all") return "all devices";
  return `${scope.kind} ${scope.id}`;
}

export function scopeLabel(scope: ControlScope): string {
  const words = scopeWords(scope);
  return `${words.charAt(0).toUpperCase()}${words.slice(1)}`;
}

export function scopeKey(scope: ControlScope): string {
  return `${scope.kind}:${scope.id}`;
}

function approveWords(request: Extract<ControlRequest, { op: "approve_mode" }>): string {
  const scope = scopeWords(request.scope);
  return request.enabled ? `Approve each command on ${scope}` : `Stop approving each command on ${scope}`;
}

export function requestWords(request: ControlRequest): string {
  switch (request.op) {
    case "hold":
      return `Hold on ${scopeWords(request.scope)}`;
    case "release":
      return `Release of ${scopeWords(request.scope)}`;
    case "estop":
      return `Emergency stop on ${scopeWords(request.scope)}`;
    case "approve_mode":
      return approveWords(request);
    case "approve":
    case "deny":
      return `${opWords[request.op]} ${request.commandId}`;
    case "state":
      return "Reading the state";
  }
}

function safeStopWords(reply: ControlReply): string {
  const stopped =
    reply.stopped.length === 0 ? "No safe stop was sent." : `Safe stop sent to ${reply.stopped.join(", ")}.`;
  const verb = reply.skipped.length === 1 ? "declares" : "declare";
  const skipped = reply.skipped.length === 0 ? "" : ` ${reply.skipped.join(", ")} ${verb} no safe stop.`;
  return `${stopped}${skipped} The scope stays held until it is released.`;
}

function replyLine(request: ControlRequest, reply: ControlReply): OutcomeLine {
  if (reply.result === "refused") {
    return { tone: "refused", text: `${requestWords(request)} was refused: ${refusalWords(reply.reason)}.` };
  }
  const applied = `${requestWords(request)} was applied.`;
  return { tone: "ok", text: request.op === "estop" ? `${applied} ${safeStopWords(reply)}` : applied };
}

function failureLine(request: ControlRequest, kind: ControlFailureKind): OutcomeLine {
  return { tone: "failed", text: `${requestWords(request)}: ${failureWords[kind]}` };
}

const checkWords: Readonly<Record<StateCheck, string>> = {
  none: "",
  pending: " Reading the node's state again…",
  read: " The state shown was read from the node just now.",
  failed: " The node's state could not be read again, so the state shown may be out of date.",
};

export type ControlResult =
  | { readonly kind: "reply"; readonly reply: ControlReply }
  | { readonly kind: "failed"; readonly failure: ControlFailureKind };

export function outcomeLine(request: ControlRequest, result: ControlResult, check: StateCheck): OutcomeLine {
  const line = result.kind === "reply" ? replyLine(request, result.reply) : failureLine(request, result.failure);
  return { ...line, text: `${line.text}${checkWords[check]}` };
}

export function pendingTitle(pending: PendingApproval): string {
  return `${pending.command} on ${pending.device}`;
}

function remainingWords(seconds: number, suffix: string): string {
  const whole = Math.round(seconds);
  if (whole <= 0) return "deadline reached; the node settles it at its next step";
  return `expires in ${durationOf(whole)}${suffix}`;
}

export function deadlineWords(pending: PendingApproval, now: NodeNow): string {
  if (pending.clock === "world") return remainingWords((pending.deadlineNs - now.worldNs) / 1e9, " (run time)");
  return remainingWords((pending.deadlineNs / 1e6 - now.wallMs) / 1000, "");
}

export function waitingWords(count: number): string {
  return count === 0 ? "No commands wait for approval." : `${counted(count, "command")} waiting for approval`;
}
