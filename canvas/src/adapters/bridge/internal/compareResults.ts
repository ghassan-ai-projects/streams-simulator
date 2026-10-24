import type {
  AlarmStats,
  ComparedAlarm,
  ComparedAlarmSystem,
  ComparedKpi,
  KpiStats,
  Presence,
  Scope,
  SystemCounts,
} from "../../../model/review";
import { isRecord, record } from "../../../model/shared/json";

const presences: readonly Presence[] = ["both", "only_in_a", "only_in_b"];

function finite(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function length(value: unknown): number | null | undefined {
  return value === null ? null : (finite(value) ?? undefined);
}

function scopeOf(value: unknown): Scope | null {
  const scope = record(value);
  const aLengthNs = length(scope.a_simulated_length_ns);
  const bLengthNs = length(scope.b_simulated_length_ns);
  return aLengthNs === undefined || bLengthNs === undefined ? null : { aLengthNs, bLengthNs };
}

function statsOf(value: unknown): KpiStats | null | undefined {
  if (value === null) return null;
  const stats = record(value);
  const mean = finite(stats.mean);
  const min = finite(stats.min);
  const max = finite(stats.max);
  const samples = finite(stats.samples);
  if (mean === null || min === null || max === null || samples === null) return undefined;
  return { mean, min, max, samples };
}

function deltaMean(value: unknown, nullable: boolean): number | null | undefined {
  if (value === undefined) return null;
  const mean = record(value).mean;
  if (mean === null && nullable) return null;
  return finite(mean) ?? undefined;
}

export function kpiRow(value: unknown): ComparedKpi | null {
  const row = record(value);
  const presence = presences.find((known) => known === row.presence);
  const [a, b] = [statsOf(row.a), statsOf(row.b)];
  const [delta, relative] = [deltaMean(row.delta, false), deltaMean(row.relative_delta, true)];
  const scope = scopeOf(row.scope);
  if (typeof row.entity_id !== "string" || typeof row.kpi !== "string" || presence === undefined) return null;
  if (a === undefined || b === undefined || delta === undefined || relative === undefined || scope === null)
    return null;
  return { entity: row.entity_id, kpi: row.kpi, presence, a, b, delta, relative, scope };
}

function alarmStats(value: unknown): AlarmStats | null | undefined {
  if (value === null) return null;
  const stats = record(value);
  const [activations, activeS] = [finite(stats.activations), finite(stats.active_s)];
  if (activations === null || activeS === null || typeof stats.priority !== "string") return undefined;
  return { priority: stats.priority, activations, activeS };
}

function alarmDelta(value: unknown): ComparedAlarm["delta"] | undefined {
  if (value === undefined) return null;
  const [activations, activeS] = [finite(record(value).activations), finite(record(value).active_s)];
  return activations === null || activeS === null ? undefined : { activations, activeS };
}

export function alarmRow(value: unknown): ComparedAlarm | null {
  const row = record(value);
  const presence = presences.find((known) => known === row.presence);
  const [a, b, delta, scope] = [alarmStats(row.a), alarmStats(row.b), alarmDelta(row.delta), scopeOf(row.scope)];
  if (typeof row.entity_id !== "string" || typeof row.alarm !== "string" || presence === undefined) return null;
  if (a === undefined || b === undefined || delta === undefined || scope === null) return null;
  return { entity: row.entity_id, alarm: row.alarm, presence, a, b, delta, scope };
}

function counts(value: unknown): SystemCounts | null | undefined {
  if (value === undefined || value === null) return value === null ? null : undefined;
  if (!isRecord(value)) return undefined;
  const numbers = Object.entries(value).filter((entry): entry is [string, number] => finite(entry[1]) !== null);
  return Object.fromEntries(numbers);
}

export function alarmSystemOf(value: unknown): ComparedAlarmSystem | null {
  const system = record(value);
  const [a, b, scope] = [counts(system.a), counts(system.b), scopeOf(system.scope)];
  const delta = system.delta === undefined ? null : counts(system.delta);
  if (a === undefined || b === undefined || delta === undefined || scope === null) return null;
  return { a, b, delta, scope };
}
