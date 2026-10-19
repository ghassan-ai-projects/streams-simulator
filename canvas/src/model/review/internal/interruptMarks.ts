import { requestWords } from "./interruptWords";
import type { ControlRequest, InterruptEvent, InterruptRecord } from "./interrupts";
import { refusalWords } from "./sessionWords";
import type { EventMark } from "./timeline";

function mark(atNs: number, label: string, raw: string): EventMark {
  return { atNs, label, raw, kind: "interrupt", tone: "other", entity: "", untilNs: null };
}

function modeRequest(record: InterruptRecord): ControlRequest | null {
  const { scope, enabled } = record;
  return scope === null || enabled === null ? null : { op: "approve_mode", scope, enabled };
}

function requestOf(record: InterruptRecord): ControlRequest | null {
  const { op, scope, commandId } = record;
  if (op === "approve" || op === "deny") return commandId === "" ? null : { op, commandId };
  if (op === "approve_mode") return modeRequest(record);
  if (op === "hold" || op === "release" || op === "estop") return scope === null ? null : { op, scope };
  return null;
}

function controlMark(record: InterruptRecord, atNs: number): EventMark | null {
  const request = requestOf(record);
  return request === null ? null : mark(atNs, `Operator: ${requestWords(request)}`, `control ${request.op}`);
}

function replyMark(record: InterruptRecord, atNs: number): EventMark | null {
  if (record.result !== "refused") return null;
  const why = refusalWords(record.reason);
  return mark(atNs, `Control refused: ${why}`, `reply refused ${record.op ?? ""} ${record.reason}`);
}

function resolveMark(record: InterruptRecord, atNs: number): EventMark {
  if (record.resolution === "forward") {
    return mark(atNs, `${record.commandId} approved and handed on`, `resolve forward ${record.commandId}`);
  }
  const why = refusalWords(record.reason);
  return mark(atNs, `${record.commandId} refused: ${why}`, `resolve refuse ${record.commandId} ${record.reason}`);
}

function stopMark(record: InterruptRecord, atNs: number): EventMark {
  return mark(atNs, `Safe stop sent to ${record.device}`, `stop ${record.device} ${record.command}`);
}

type Maker = (record: InterruptRecord, atNs: number) => EventMark | null;

const makers: Readonly<Partial<Record<InterruptEvent, Maker>>> = {
  control: controlMark,
  reply: replyMark,
  resolve: resolveMark,
  stop: stopMark,
};

export function interruptMarks(record: InterruptRecord, atNs: number): readonly EventMark[] {
  const made = makers[record.event]?.(record, atNs) ?? null;
  return made === null ? [] : [made];
}
