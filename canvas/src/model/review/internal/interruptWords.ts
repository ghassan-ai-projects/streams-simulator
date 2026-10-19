import { durationOf } from "../../shared/duration";
import { counted } from "../../shared/words";
import { coverAround } from "./interruptCover";
import { scopeLabel, scopeWords } from "./interruptScope";
import type {
  ActionRequest,
  ControlFailureKind,
  ControlReply,
  ControlScope,
  InterruptMode,
  PendingApproval,
  WordedRequest,
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

const modeText = {
  hold: "Held",
  latched: "Held · latched by emergency stop",
  approve: "Approving each command",
} as const;

export type BadgeKind = keyof typeof modeText;

const decisionWords: Readonly<Record<"approve" | "deny", string>> = { approve: "Approve", deny: "Deny" };

const failureWords: Readonly<Record<ControlFailureKind, string>> = {
  "no-session": "No connected session is running, so nothing was sent.",
  "no-controls": "This session has no control socket, so the node cannot be controlled.",
  gone: "The node's control socket is not there. Nothing was confirmed.",
  timeout: "The node did not answer in time. The control may still apply afterwards.",
  unavailable: "The bridge did not answer. Nothing was confirmed.",
  unreadable: "The answer could not be read. Nothing was confirmed.",
  rejected: "The bridge refused the request, so nothing was confirmed.",
};

interface ModeFacts {
  readonly mode: InterruptMode;
  readonly latched: boolean;
}

export function badgeKind({ mode, latched }: ModeFacts): BadgeKind {
  if (mode === "approve") return "approve";
  return latched ? "latched" : "hold";
}

export function modeLine(facts: ModeFacts): string {
  return modeText[badgeKind(facts)];
}

function approveWords(request: Extract<ActionRequest, { op: "approve_mode" }>): string {
  const scope = scopeWords(request.scope);
  return request.enabled ? `Approve each command on ${scope}` : `Stop approving each command on ${scope}`;
}

export function requestWords(request: WordedRequest): string {
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
      return `${decisionWords[request.op]} ${request.commandId}`;
  }
}

function listed(names: readonly string[]): string {
  return names.join(", ");
}

function safeStopWords(reply: ControlReply): string {
  const stopped = reply.stopped.length === 0 ? "No safe stop was sent." : `Safe stop sent to ${listed(reply.stopped)}.`;
  const verb = reply.skipped.length === 1 ? "declares" : "declare";
  const skipped = reply.skipped.length === 0 ? "" : ` ${listed(reply.skipped)} ${verb} no safe stop.`;
  const unreached =
    reply.unreachable.length === 0
      ? ""
      : ` The stop could NOT be sent to ${listed(reply.unreachable)}: they may still be energised.`;
  return `${stopped}${skipped}${unreached} The scope stays held until it is released.`;
}

function narrowerWords(narrower: readonly ControlScope[]): string {
  if (narrower.length === 0) return "";
  const verb = narrower.length === 1 ? "remains" : "remain";
  return `${counted(narrower.length, "narrower hold")} ${verb}: ${narrower.map(scopeWords).join(", ")}.`;
}

function widerWords(scope: ControlScope, wider: ControlScope | null): string {
  return wider === null ? "" : `${scopeLabel(scope)} is still held through ${scopeWords(wider)}.`;
}

function releasedWords(scope: ControlScope, reply: ControlReply): string {
  const cover = coverAround(
    scope,
    reply.state.holds.map((hold) => hold.scope),
    reply.state,
  );
  const rest = [narrowerWords(cover.narrower), widerWords(scope, cover.wider)].filter((part) => part !== "");
  return [`Released ${scopeWords(scope)}.`, ...rest].join(" ");
}

function appliedLine(request: ActionRequest, reply: ControlReply): OutcomeLine {
  if (request.op === "release") return { tone: "ok", text: releasedWords(request.scope, reply) };
  const applied = `${requestWords(request)} was applied.`;
  if (request.op !== "estop") return { tone: "ok", text: applied };
  return { tone: reply.unreachable.length > 0 ? "failed" : "ok", text: `${applied} ${safeStopWords(reply)}` };
}

function replyLine(request: ActionRequest, reply: ControlReply): OutcomeLine {
  if (reply.result === "refused") {
    return { tone: "refused", text: `${requestWords(request)} was refused: ${refusalWords(reply.reason)}.` };
  }
  return appliedLine(request, reply);
}

function failureLine(request: ActionRequest, kind: ControlFailureKind, detail: string): OutcomeLine {
  const why = detail === "" ? "" : ` (${detail})`;
  return { tone: "failed", text: `${requestWords(request)}: ${failureWords[kind]}${why}` };
}

const checkWords: Readonly<Record<StateCheck, string>> = {
  none: "",
  pending: " Reading the node's state again…",
  read: " The state shown was read from the node just now.",
  failed: " The node's state could not be read again, so the state shown may be out of date.",
};

export type ControlResult =
  | { readonly kind: "reply"; readonly reply: ControlReply }
  | { readonly kind: "failed"; readonly failure: ControlFailureKind; readonly detail: string }
  | { readonly kind: "not-sent" };

function resultLine(request: ActionRequest, result: ControlResult): OutcomeLine {
  if (result.kind === "reply") return replyLine(request, result.reply);
  if (result.kind === "not-sent") {
    return { tone: "refused", text: `${requestWords(request)}: not sent: another control is still waiting.` };
  }
  return failureLine(request, result.failure, result.detail);
}

export function outcomeLine(request: ActionRequest, result: ControlResult, check: StateCheck): OutcomeLine {
  const line = resultLine(request, result);
  return { ...line, text: `${line.text}${checkWords[check]}` };
}

export function pendingTitle(pending: PendingApproval): string {
  return `${pending.command} on ${pending.device}`;
}

const shownArgs = 6;
const shownValue = 40;

function valueWords(value: string): string {
  return value.length > shownValue ? `${value.slice(0, shownValue)}…` : value;
}

export function argsWords(pending: PendingApproval): string {
  if (pending.argsTruncated) return "arguments too large to show";
  if (pending.args.length === 0) return "no arguments";
  const shown = pending.args.slice(0, shownArgs).map((pair) => `${pair.name}: ${valueWords(pair.value)}`);
  const more = pending.args.length - shownArgs;
  return more > 0 ? `${shown.join(", ")} and ${String(more)} more` : shown.join(", ");
}

export function arrivalWords(pending: PendingApproval): string {
  return `A command is waiting for approval: ${pendingTitle(pending)} (${pending.commandId}).`;
}

function remainingWords(seconds: number, suffix: string): string {
  const whole = Math.round(seconds);
  if (whole <= 0) return "deadline reached; the node settles it at its next step";
  return `expires in about ${durationOf(whole)}${suffix}`;
}

export function deadlineWords(pending: PendingApproval, now: NodeNow): string {
  if (pending.clock === "world") return remainingWords((pending.deadlineNs - now.worldNs) / 1e9, " (run time)");
  return remainingWords((pending.deadlineNs / 1e6 - now.wallMs) / 1000, "");
}

export function waitingWords(count: number): string {
  return count === 0 ? "No commands wait for approval." : `${counted(count, "command")} waiting for approval`;
}
