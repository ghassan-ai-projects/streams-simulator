import type {
  ControlOp,
  ControlReply,
  ControlScope,
  CoveredDevice,
  InterruptEvent,
  InterruptHold,
  InterruptRecord,
  InterruptState,
  PendingApproval,
} from "../../../model/review";
import { isRecord, number, text } from "../../../model/shared/json";
import { names, pick, rows } from "./decodeKit";

const scopeKinds = ["all", "board", "device"] as const;
const clocks = ["world", "wall"] as const;
const modes = ["hold", "approve"] as const;
const results = ["applied", "refused"] as const;
const verdicts = ["pass", "refuse", "queue"] as const;
const resolutions = ["forward", "refuse"] as const;
const controlOps: readonly ControlOp[] = ["hold", "release", "approve_mode", "approve", "deny", "estop", "state"];
const interruptEvents: readonly InterruptEvent[] = [
  "inventory",
  "control",
  "command",
  "clock",
  "flush",
  "reply",
  "verdict",
  "resolve",
  "stop",
  "effect",
];

interface Given<T> {
  readonly value: T;
}

function optional<T extends string>(value: unknown, choices: readonly T[]): Given<T | null> | null {
  if (value === undefined) return { value: null };
  const found = pick(value, choices);
  return found === null ? null : { value: found };
}

function scopeOf(row: Record<string, unknown>): ControlScope | null {
  const kind = pick(row.kind, scopeKinds);
  const id = text(row.id);
  if (kind === null || (row.id !== undefined && typeof row.id !== "string")) return null;
  return (kind === "all") === (id === "") ? { kind, id } : null;
}

function optionalScope(value: unknown): Given<ControlScope | null> | null {
  if (value === undefined) return { value: null };
  const scope = isRecord(value) ? scopeOf(value) : null;
  return scope === null ? null : { value: scope };
}

function holdOf(row: Record<string, unknown>): InterruptHold | null {
  const scope = scopeOf(row);
  return scope === null || typeof row.latched !== "boolean" ? null : { scope, latched: row.latched };
}

function pendingOf(row: Record<string, unknown>): PendingApproval | null {
  const clock = pick(row.clock, clocks);
  const numbers = typeof row.deadline_ns === "number" && typeof row.queued_at_ns === "number";
  if (clock === null || !numbers || text(row.command_id) === "") return null;
  return {
    commandId: text(row.command_id),
    device: text(row.device),
    command: text(row.command),
    deadlineNs: number(row.deadline_ns),
    queuedAtNs: number(row.queued_at_ns),
    clock,
  };
}

function coveredOf(row: Record<string, unknown>): CoveredDevice | null {
  const mode = pick(row.mode, modes);
  if (mode === null || typeof row.latched !== "boolean" || text(row.device) === "") return null;
  return { device: text(row.device), board: text(row.board), mode, latched: row.latched };
}

function interruptState(value: unknown): InterruptState | null {
  if (!isRecord(value)) return null;
  const holds = rows(value.holds, holdOf);
  const approve = rows(value.approve, scopeOf);
  const pending = rows(value.pending, pendingOf);
  const devices = rows(value.devices, coveredOf);
  if (holds === null || approve === null || pending === null || devices === null) return null;
  return { holds, approve, pending, devices };
}

export function optionalInterrupts(value: unknown): Given<InterruptState | null> | null {
  if (value === undefined || value === null) return { value: null };
  const state = interruptState(value);
  return state === null ? null : { value: state };
}

interface Variants {
  readonly op: ControlOp | null;
  readonly result: "applied" | "refused" | null;
  readonly verdict: "pass" | "refuse" | "queue" | null;
  readonly resolution: "forward" | "refuse" | null;
}

function variantsOf(row: Record<string, unknown>): Variants | null {
  const op = optional(row.op, controlOps);
  const result = optional(row.result, results);
  const verdict = optional(row.verdict, verdicts);
  const resolution = optional(row.resolution, resolutions);
  if (op === null || result === null || verdict === null || resolution === null) return null;
  return { op: op.value, result: result.value, verdict: verdict.value, resolution: resolution.value };
}

export function interruptRecord(row: Record<string, unknown>): InterruptRecord | null {
  const event = pick(row.event, interruptEvents);
  const clock = pick(row.clock, clocks);
  const variants = variantsOf(row);
  const scope = optionalScope(row.scope);
  const flagged = row.enabled === undefined || typeof row.enabled === "boolean";
  if (event === null || clock === null || variants === null || scope === null || !flagged) return null;
  if (typeof row.node_ns !== "number") return null;
  return {
    event,
    clock,
    ...variants,
    nodeNs: row.node_ns,
    scope: scope.value,
    enabled: typeof row.enabled === "boolean" ? row.enabled : null,
    commandId: text(row.command_id),
    device: text(row.device),
    command: text(row.command),
    reason: text(row.reason),
  };
}

export function controlReply(value: unknown): ControlReply | null {
  if (!isRecord(value)) return null;
  const result = pick(value.result, results);
  const op = optional(value.op, controlOps);
  const state = interruptState(value.state);
  const stopped = names(value.stopped);
  const skipped = names(value.skipped);
  if (result === null || op === null || state === null || stopped === null || skipped === null) return null;
  return {
    result,
    op: op.value,
    reason: text(value.reason),
    commandId: text(value.command_id),
    stopped,
    skipped,
    state,
  };
}
