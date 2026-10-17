import type { JournalEntry } from "../../../model/review";
import { isRecord, string } from "./http";

const journalTypes = [
  "observation",
  "run.opened",
  "run.closed",
  "clock.advanced",
  "effector.acknowledged",
  "consumer.reported",
  "wire",
  "pacing",
  "command",
  "delivery",
  "state",
  "effector.outcome",
  "kpi.truth",
  "fault.truth",
  "run.unblinded",
] as const;

export function journalEntry(value: string): JournalEntry | null {
  const parsed = parseJson(value);
  if (!isRecord(parsed) || !validEnvelope(parsed) || !validEvent(parsed)) return null;
  return projectEntry(parsed);
}

function validEnvelope(record: Record<string, unknown>): boolean {
  return (
    journalTypes.includes(record.type as (typeof journalTypes)[number]) &&
    typeof record.at_ns === "number" &&
    Number.isFinite(record.at_ns) &&
    typeof record.seq === "number" &&
    Number.isInteger(record.seq) &&
    (record.stream === "observer" || record.stream === "director")
  );
}

function validEvent(record: Record<string, unknown>): boolean {
  if (record.type === "observation") return typeof record.line === "string";
  if (record.type === "run.opened") return openedRecord(record);
  return record.type !== "run.closed" || typeof record.incomplete === "boolean";
}

function openedRecord(record: Record<string, unknown>): boolean {
  return typeof record.mode === "string" && string(object(record.adapter).id) !== "";
}

function projectEntry(record: Record<string, unknown>): JournalEntry {
  const atNs = number(record.at_ns);
  if (record.type === "observation") return observation(record, atNs);
  if (record.type === "run.opened") {
    return { kind: "opened", atNs, mode: string(record.mode), adapter: string(object(record.adapter).id) };
  }
  if (record.type === "run.closed") return { kind: "closed", atNs, incomplete: record.incomplete === true };
  return { kind: "other", atNs };
}

function observation(record: Record<string, unknown>, atNs: number): JournalEntry {
  const line = object(parseJson(string(record.line)));
  const raw = line.value;
  return {
    kind: "observation",
    observation: {
      atNs,
      entity: string(line.entity_id),
      channel: string(line.channel),
      value: typeof raw === "number" || typeof raw === "boolean" ? raw : string(raw),
      unit: string(line.unit),
    },
  };
}

function parseJson(value: string): unknown {
  try {
    return JSON.parse(value) as unknown;
  } catch {
    return null;
  }
}

function object(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {};
}

function number(value: unknown): number {
  return typeof value === "number" ? value : 0;
}
