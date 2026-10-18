import type { LiveMark, SessionLive, SessionDevice, SessionCommand, SessionReading } from "../../../model/review";
import { isRecord, list, number, record, text } from "../../../model/shared/json";

const environments = ["simulated", "hardware", "unknown"] as const;
const links = ["unknown", "online", "offline", "not_applicable"] as const;

function choice<T extends string>(value: unknown, choices: readonly T[]): value is T {
  return typeof value === "string" && choices.includes(value as T);
}

function rows<T>(value: unknown, decode: (row: Record<string, unknown>) => T | null): readonly T[] | null {
  if (!Array.isArray(value)) return null;
  const result = value.map((row) => (isRecord(row) ? decode(row) : null));
  return result.includes(null) ? null : (result as T[]);
}

function reading(row: Record<string, unknown>): SessionReading | null {
  if (!choice(row.state, ["good", "simulated", "stale", "disconnected", "invalid", "untrusted"])) return null;
  const scalar = typeof row.value === "number" || typeof row.value === "string" || typeof row.value === "boolean";
  if (row.state === "good" && !scalar) return null;
  return {
    name: text(row.name),
    unit: text(row.unit),
    value: scalar ? (row.value as number | string | boolean) : null,
    quality: text(row.quality),
    state: row.state,
  };
}

function device(row: Record<string, unknown>): SessionDevice | null {
  if (!choice(row.environment, environments) || !choice(row.health, ["online", "offline", "simulated"])) return null;
  const readings = rows(row.readings, reading);
  if (readings === null) return null;
  return {
    device: text(row.device),
    board: text(row.board),
    instance: text(row.instance),
    environment: row.environment,
    bootId: text(row.boot_id),
    health: row.health,
    reboots: number(row.reboots),
    readings,
  };
}

function stagesHold(row: Record<string, unknown>): boolean {
  if (row.sent !== true) return false;
  if (row.effect === "unknown") return true;
  return (
    row.effect === "observed" &&
    row.receipt === "accepted" &&
    row.report === "succeeded" &&
    text(row.effect_evidence) !== ""
  );
}

function command(row: Record<string, unknown>): SessionCommand | null {
  if (
    !choice(row.receipt, ["pending", "accepted", "refused"]) ||
    !choice(row.report, ["pending", "succeeded", "failed", "unknown"]) ||
    !choice(row.effect, ["unknown", "observed"]) ||
    !choice(row.environment, environments) ||
    !stagesHold(row)
  )
    return null;
  return {
    commandId: text(row.command_id),
    device: text(row.device),
    environment: row.environment,
    command: text(row.command),
    bootId: text(row.boot_id),
    receipt: row.receipt,
    reason: text(row.reason),
    report: row.report,
    effect: row.effect,
    effectEvidence: text(row.effect_evidence),
    sentAtNs: number(row.sent_at_ns),
    updatedAtNs: number(row.updated_at_ns),
  };
}

function mark(row: Record<string, unknown>): LiveMark | null {
  if (!choice(row.kind, ["command", "connection"]) || !choice(row.subject, ["host", "gateway", "device", "command"]))
    return null;
  return {
    atNs: number(row.at_ns),
    kind: row.kind,
    subject: row.subject,
    label: text(row.label),
    entity: text(row.entity),
  };
}

function commandState(row: unknown): Record<string, unknown> {
  return Object.fromEntries(Object.entries(record(row)).filter(([name]) => name !== "updated_at_ns"));
}

function stateKey(row: Record<string, unknown>): string {
  const commands = list(row.commands).map(commandState);
  return JSON.stringify([row.host, row.gateway, row.devices, commands, row.commands_truncated]);
}

export function liveSnapshot(row: Record<string, unknown>): SessionLive | null {
  if (!choice(row.host, links) || !choice(row.gateway, links)) return null;
  const devices = rows(row.devices, device);
  const commands = rows(row.commands, command);
  const marks = rows(row.marks, mark);
  if (devices === null || commands === null || marks === null) return null;
  return {
    key: stateKey(row),
    atNs: number(row.at_ns),
    host: row.host,
    gateway: row.gateway,
    devices,
    commands,
    marks,
    commandsTruncated: row.commands_truncated === true,
  };
}
