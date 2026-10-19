import type {
  ArgPair,
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
import { isRecord } from "../../../model/shared/json";
import type { Given } from "./decodeKit";
import { names, optional, optionalTexts, pick, requiredTexts, rows } from "./decodeKit";

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
  "purge",
];

function scopeOf(row: Record<string, unknown>): ControlScope | null {
  const kind = pick(row.kind, scopeKinds);
  const named = optionalTexts(row, ["id"]);
  if (kind === null || named === null) return null;
  return (kind === "all") === (named.id === "") ? { kind, id: named.id } : null;
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

function argText(value: unknown): string {
  return typeof value === "string" ? value : JSON.stringify(value);
}

function argPairs(value: unknown): readonly ArgPair[] | null {
  if (!isRecord(value)) return null;
  return Object.entries(value).map(([name, entry]) => ({ name, value: argText(entry) }));
}

function pendingOf(row: Record<string, unknown>): PendingApproval | null {
  const clock = pick(row.clock, clocks);
  const ids = requiredTexts(row, ["command_id", "device", "command", "content_digest"] as const);
  const args = argPairs(row.args);
  const deadlineNs = row.deadline_ns;
  const queuedAtNs = row.queued_at_ns;
  if (clock === null || ids === null || args === null) return null;
  if (typeof deadlineNs !== "number" || typeof queuedAtNs !== "number") return null;
  if (typeof row.args_truncated !== "boolean") return null;
  return {
    commandId: ids.command_id,
    device: ids.device,
    command: ids.command,
    contentDigest: ids.content_digest,
    deadlineNs,
    queuedAtNs,
    clock,
    args,
    argsTruncated: row.args_truncated,
  };
}

function coveredOf(row: Record<string, unknown>): CoveredDevice | null {
  const mode = pick(row.mode, modes);
  const device = requiredTexts(row, ["device"] as const);
  const board = optionalTexts(row, ["board"] as const);
  if (mode === null || typeof row.latched !== "boolean" || device === null || board === null) return null;
  return { device: device.device, board: board.board, mode, latched: row.latched };
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

interface Head {
  readonly event: InterruptEvent;
  readonly clock: "world" | "wall";
  readonly nodeNs: number;
  readonly scope: ControlScope | null;
  readonly enabled: boolean | null;
}

function headOf(row: Record<string, unknown>): (Head & Variants) | null {
  const event = pick(row.event, interruptEvents);
  const clock = pick(row.clock, clocks);
  const variants = variantsOf(row);
  const scope = optionalScope(row.scope);
  const flagged = row.enabled === undefined || typeof row.enabled === "boolean";
  if (event === null || clock === null || variants === null || scope === null || !flagged) return null;
  if (typeof row.node_ns !== "number") return null;
  const enabled = typeof row.enabled === "boolean" ? row.enabled : null;
  return { ...variants, event, clock, nodeNs: row.node_ns, scope: scope.value, enabled };
}

export function interruptRecord(row: Record<string, unknown>): InterruptRecord | null {
  const head = headOf(row);
  const given = optionalTexts(row, ["command_id", "device", "command", "reason"] as const);
  if (head === null || given === null) return null;
  return {
    ...head,
    commandId: given.command_id,
    device: given.device,
    command: given.command,
    reason: given.reason,
  };
}

export function controlReply(value: unknown): ControlReply | null {
  if (!isRecord(value)) return null;
  const result = pick(value.result, results);
  const op = optional(value.op, controlOps);
  const state = interruptState(value.state);
  const lists = [names(value.stopped), names(value.skipped), names(value.unreachable)] as const;
  const given = optionalTexts(value, ["reason", "command_id"] as const);
  if (result === null || op === null || state === null || given === null) return null;
  const [stopped, skipped, unreachable] = lists;
  if (stopped === null || skipped === null || unreachable === null) return null;
  return {
    result,
    op: op.value,
    reason: given.reason,
    commandId: given.command_id,
    stopped,
    skipped,
    unreachable,
    state,
  };
}
