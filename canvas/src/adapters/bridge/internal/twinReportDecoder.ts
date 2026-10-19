import type {
  TwinCommand,
  TwinDivergence,
  TwinMetric,
  TwinReport,
  TwinSeries,
  TwinSummary,
  TwinTolerance,
} from "../../../model/review";
import {
  commandStatuses,
  notAppliedReasons,
  notComparedReasons,
  toleranceSources,
  twinJudgeReasons,
  twinVerdicts,
} from "../../../model/review";
import { isRecord, text } from "../../../model/shared/json";
import { names, optional, pick, rows } from "./decodeKit";

type Wire = Record<string, unknown>;

function finite(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value);
}

function numbersOf<K extends string>(value: unknown, keys: readonly K[]): Record<K, number> | null {
  if (!isRecord(value)) return null;
  const entries = keys.map((key) => [key, value[key]] as const);
  return entries.every(([, entry]) => finite(entry)) ? (Object.fromEntries(entries) as Record<K, number>) : null;
}

function named(row: Wire, keys: readonly string[]): boolean {
  return keys.every((key) => text(row[key]) !== "");
}

function summaryOf(value: unknown): TwinSummary | null {
  const n = numbersOf(value, [
    "devices",
    "metrics",
    "compared",
    "within",
    "diverged",
    "not_compared",
    "unattributed_readings",
    "metrics_cannot_judge",
    "divergences",
    "commands_applied",
    "commands_not_applied",
  ] as const);
  if (n === null) return null;
  return {
    devices: n.devices,
    metrics: n.metrics,
    compared: n.compared,
    within: n.within,
    diverged: n.diverged,
    notCompared: n.not_compared,
    unattributed: n.unattributed_readings,
    metricsCannotJudge: n.metrics_cannot_judge,
    divergences: n.divergences,
    commandsApplied: n.commands_applied,
    commandsNotApplied: n.commands_not_applied,
  };
}

function commandOf(row: Wire): TwinCommand | null {
  const status = pick(row.status, commandStatuses);
  const reason = optional(row.reason, notAppliedReasons);
  const timed = row.at_ns === undefined || finite(row.at_ns);
  if (status === null || reason === null || !timed || !named(row, ["command_id", "device", "command"])) return null;
  if (status === "not_applied" && reason.value === null) return null;
  return {
    commandId: text(row.command_id),
    device: text(row.device),
    command: text(row.command),
    atNs: finite(row.at_ns) ? row.at_ns : null,
    status,
    reason: reason.value,
    detail: text(row.detail),
  };
}

function toleranceOf(value: unknown): TwinTolerance | null {
  if (!isRecord(value)) return null;
  const source = pick(value.source, toleranceSources);
  const amount = value.value;
  if (amount !== null && !(finite(amount) && amount >= 0)) return null;
  return source === null ? null : { value: amount, source };
}

function seriesOf(value: unknown): TwinSeries | null {
  if (!isRecord(value) || !finite(value.total) || typeof value.truncated !== "boolean") return null;
  const points = rows(value.points, (row) => {
    const n = numbersOf(row, ["at_ns", "board", "twin"] as const);
    return n === null ? null : { atNs: n.at_ns, board: n.board, twin: n.twin };
  });
  return points === null ? null : { points, total: value.total, truncated: value.truncated };
}

function metricOf(row: Wire): TwinMetric | null {
  const verdict = pick(row.verdict, twinVerdicts);
  const tolerance = toleranceOf(row.tolerance);
  const counts = numbersOf(row.summary, ["compared", "within", "diverged", "not_compared"] as const);
  const reasons = rows(row.not_compared_reasons, (entry) => {
    const reason = pick(entry.reason, notComparedReasons);
    return reason === null || !finite(entry.count) ? null : { reason, count: entry.count };
  });
  const series = seriesOf(row.series);
  if (verdict === null || tolerance === null || counts === null || reasons === null || series === null) return null;
  if (!named(row, ["device", "metric"])) return null;
  const { compared, within, diverged } = counts;
  return {
    device: text(row.device),
    kind: text(row.kind),
    metric: text(row.metric),
    unit: text(row.unit),
    tolerance,
    verdict,
    counts: { compared, within, diverged, notCompared: counts.not_compared },
    notCompared: reasons,
    series,
  };
}

function divergenceOf(row: Wire): TwinDivergence | null {
  const n = numbersOf(row, [
    "first_ns",
    "last_ns",
    "readings",
    "tolerance",
    "worst_error",
    "worst_at_ns",
    "board_value",
    "twin_value",
  ] as const);
  const commands = names(row.commands);
  if (n === null || commands === null || !named(row, ["name", "device", "metric"])) return null;
  return {
    name: text(row.name),
    device: text(row.device),
    metric: text(row.metric),
    firstNs: n.first_ns,
    lastNs: n.last_ns,
    readings: n.readings,
    tolerance: n.tolerance,
    worstError: n.worst_error,
    worstAtNs: n.worst_at_ns,
    boardValue: n.board_value,
    twinValue: n.twin_value,
    commands,
  };
}

function headOf(value: Wire) {
  const verdict = pick(value.verdict, twinVerdicts);
  const reason = optional(value.reason, twinJudgeReasons);
  const summary = summaryOf(value.summary);
  if (verdict === null || reason === null || summary === null) return null;
  return { verdict, reason: reason.value, summary };
}

function bodyOf(value: Wire) {
  const twin = isRecord(value.twin) && text(value.twin.domain) !== "" ? text(value.twin.domain) : null;
  const rule: Wire = isRecord(value.tolerance_rule) ? value.tolerance_rule : {};
  const commands = rows(value.commands, commandOf);
  const metrics = rows(value.metrics, metricOf);
  const divergences = rows(value.divergences, divergenceOf);
  if (twin === null || !finite(rule.span_fraction) || commands === null) return null;
  if (metrics === null || divergences === null) return null;
  return {
    twinDomain: twin,
    spanFraction: rule.span_fraction,
    ruleDescription: text(rule.description),
    commands,
    metrics,
    divergences,
  };
}

export function twinReport(value: unknown): TwinReport | null {
  if (!isRecord(value) || value.schema_version !== "0.1") return null;
  const head = headOf(value);
  const body = bodyOf(value);
  return head === null || body === null ? null : { ...head, ...body };
}
