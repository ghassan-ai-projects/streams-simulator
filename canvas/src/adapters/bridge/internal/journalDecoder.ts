import type { JournalEntry } from "../../../model/review";
import { isRecord, number, record, text } from "../../../model/shared/json";

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

function validEnvelope(entry: Record<string, unknown>): boolean {
  return (
    journalTypes.includes(entry.type as (typeof journalTypes)[number]) &&
    typeof entry.at_ns === "number" &&
    Number.isFinite(entry.at_ns) &&
    typeof entry.seq === "number" &&
    Number.isInteger(entry.seq) &&
    (entry.stream === "observer" || entry.stream === "director")
  );
}

function validEvent(entry: Record<string, unknown>): boolean {
  if (entry.type === "observation") return typeof entry.line === "string";
  if (entry.type === "run.opened") return openedRecord(entry);
  return entry.type !== "run.closed" || typeof entry.incomplete === "boolean";
}

function openedRecord(entry: Record<string, unknown>): boolean {
  return typeof entry.mode === "string" && text(record(entry.adapter).id) !== "";
}

function projectEntry(entry: Record<string, unknown>): JournalEntry {
  const atNs = number(entry.at_ns);
  if (entry.type === "observation") return observation(entry, atNs);
  if (entry.type === "run.opened") {
    return { kind: "opened", atNs, mode: text(entry.mode), adapter: text(record(entry.adapter).id) };
  }
  if (entry.type === "run.closed") return { kind: "closed", atNs, incomplete: entry.incomplete === true };
  return { kind: "other", atNs };
}

function observation(entry: Record<string, unknown>, atNs: number): JournalEntry {
  const line = record(parseJson(text(entry.line)));
  const raw = line.value;
  return {
    kind: "observation",
    observation: {
      atNs,
      entity: text(line.entity_id),
      channel: text(line.channel),
      value: typeof raw === "number" || typeof raw === "boolean" ? raw : text(raw),
      unit: text(line.unit),
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
