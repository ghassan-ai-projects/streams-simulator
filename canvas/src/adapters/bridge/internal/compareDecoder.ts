import type { ComparabilityLimit, ComparedRun, Comparison } from "../../../model/review";
import { isRecord, list, record, text } from "../../../model/shared/json";
import { callItem, diffOf, environmentItem, faultItem, perturbationItem } from "./compareItems";
import { alarmRow, alarmSystemOf, kpiRow } from "./compareResults";

function runOf(value: unknown): ComparedRun | null {
  const run = record(value);
  return typeof run.start_time_ns === "number"
    ? { runId: text(run.run_id), name: text(run.name), startNs: run.start_time_ns }
    : null;
}

function limitOf(value: unknown): ComparabilityLimit | null {
  const limit = record(value);
  if (typeof limit.code !== "string" || typeof limit.message !== "string") return null;
  return {
    code: limit.code,
    side: limit.side === "a" || limit.side === "b" ? limit.side : null,
    message: limit.message,
  };
}

function allOf<T>(value: unknown, item: (each: unknown) => T | null): readonly T[] | null {
  if (!Array.isArray(value)) return null;
  const read = value.map(item);
  return read.every((each) => each !== null) ? read : null;
}

function planOf(value: unknown): Comparison["plan"] | null {
  const plan = record(value);
  const faults = diffOf(plan.faults, faultItem);
  const perturbations = diffOf(plan.perturbations, perturbationItem);
  const environment = diffOf(plan.environment, environmentItem);
  const effectorCalls = diffOf(plan.effector_calls, callItem);
  if (faults === null || perturbations === null || environment === null || effectorCalls === null) return null;
  return { faults, perturbations, environment, effectorCalls };
}

type Results = Pick<Comparison, "kpis" | "alarms" | "alarmSystem">;

function resultsOf(value: unknown): Results | null {
  const results = record(value);
  const kpis = allOf(results.kpis, kpiRow);
  const alarms = allOf(results.alarms, alarmRow);
  const alarmSystem = alarmSystemOf(results.alarm_system);
  return kpis === null || alarms === null || alarmSystem === null ? null : { kpis, alarms, alarmSystem };
}

export function parseComparisonPayload(value: unknown): Comparison | null {
  if (!isRecord(value)) return null;
  const [a, b, plan] = [runOf(value.a), runOf(value.b), planOf(value.plan)];
  const limits = allOf(value.comparability_limits, limitOf);
  const results = resultsOf(value.results);
  const windowedReason = text(record(value.windowed_results).reason);
  if (a === null || b === null || plan === null || limits === null || results === null) return null;
  if (windowedReason === "" || list(value.comparability_limits).length === 0) return null;
  return { a, b, plan, limits, ...results, windowedReason };
}
