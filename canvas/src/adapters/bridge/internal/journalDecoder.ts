import type { JournalEntry } from "../../../model/review";
import { isRecord, list, number, record, text } from "../../../model/shared/json";
import { issueList } from "./http";
import { interruptRecord } from "./interruptDecoder";
import { liveSnapshot } from "./liveDecoder";

const journalTypes = [
  "live",
  "interrupt",
  "session",
  "observation",
  "run.opened",
  "run.closed",
  "clock.advanced",
  "effector.acknowledged",
  "consumer.reported",
  "wire",
  "pacing",
  "origin",
  "binding",
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
  if (parsed.type === "live") {
    const live = liveSnapshot(parsed);
    if (live === null) return { kind: "live-unreadable", atNs: number(parsed.at_ns) };
    return { kind: "live", atNs: live.atNs, live };
  }
  if (parsed.type === "interrupt") return interruptEntry(parsed);
  return projectEntry(parsed);
}

function interruptEntry(parsed: Record<string, unknown>): JournalEntry {
  const atNs = number(parsed.at_ns);
  const record = interruptRecord(parsed);
  return record === null ? { kind: "interrupt-unreadable", atNs } : { kind: "interrupt", atNs, record };
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
  if (entry.type === "session") return validSession(entry);
  if (entry.type === "run.opened") return openedRecord(entry);
  if (entry.type === "origin")
    return ["hardware", "simulated"].includes(text(entry.environment)) && text(entry.device) !== "";
  if (entry.type === "binding")
    return (entry.outcome === "matched" || entry.outcome === "refused") && Array.isArray(entry.issues);
  return entry.type !== "run.closed" || typeof entry.incomplete === "boolean";
}

function validSession(entry: Record<string, unknown>): boolean {
  return entry.mode === "connected" && Array.isArray(record(entry.plan).instances);
}

function openedRecord(entry: Record<string, unknown>): boolean {
  return typeof entry.mode === "string" && text(record(entry.adapter).id) !== "";
}

function projectEntry(entry: Record<string, unknown>): JournalEntry {
  const atNs = number(entry.at_ns);
  if (entry.type === "session") {
    const instances = list(record(entry.plan).instances).map(record);
    return {
      kind: "session",
      atNs,
      domain: text(record(entry.plan).domain),
      simulated: instances
        .filter((instance) => instance.origin === "simulated")
        .map((instance) => text(instance.instance)),
    };
  }
  if (entry.type === "observation") return observation(entry, atNs);
  if (entry.type === "run.opened") {
    return { kind: "opened", atNs, mode: text(entry.mode), adapter: text(record(entry.adapter).id) };
  }
  if (entry.type === "run.closed") return { kind: "closed", atNs, incomplete: entry.incomplete === true };
  if (entry.type === "origin") return originEntry(entry, atNs);
  if (entry.type === "binding") return bindingEntry(entry, atNs);
  return { kind: "other", atNs };
}

function originEntry(entry: Record<string, unknown>, atNs: number): JournalEntry {
  const origin = {
    instance: text(entry.instance),
    kind: entry.environment === "hardware" ? ("real" as const) : ("simulated" as const),
    board: text(entry.board),
    device: text(entry.device),
    bootId: text(entry.boot_id),
  };
  return { kind: "origin", atNs, origin };
}

function bindingEntry(entry: Record<string, unknown>, atNs: number): JournalEntry {
  const binding = {
    matched: entry.outcome === "matched",
    refusal: text(entry.refusal),
    reason: text(entry.reason),
    issues: issueList(entry.issues),
  };
  return { kind: "binding", atNs, binding };
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
