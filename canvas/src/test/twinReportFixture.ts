import type { TwinReport } from "../model/review";
import { reportWire } from "./runFixture";

const digest = `sha256:${"a".repeat(64)}`;

const points = [
  { at_ns: 100e9, board: 20, twin: 20 },
  { at_ns: 200e9, board: 21, twin: 21 },
  { at_ns: 300e9, board: 22, twin: 22 },
  { at_ns: 400e9, board: 28, twin: 22 },
  { at_ns: 500e9, board: 30, twin: 22 },
];

export function twinMetricWire(overrides: Record<string, unknown> = {}) {
  return {
    device: "sensor",
    kind: "thermal_sensor",
    metric: "temperature",
    unit: "Cel",
    tolerance: { value: 2.5, source: "default_span" },
    verdict: "diverges",
    summary: { compared: 5, within: 3, diverged: 2, not_compared: 2 },
    not_compared_reasons: [{ reason: "before_first_command", count: 2 }],
    series: { points, total: 5, truncated: false },
    ...overrides,
  };
}

function twinDivergenceWire(overrides: Record<string, unknown> = {}) {
  return {
    name: "sensor.temperature#1",
    device: "sensor",
    metric: "temperature",
    first_ns: 400e9,
    last_ns: 500e9,
    readings: 2,
    tolerance: 2.5,
    worst_error: 8,
    worst_at_ns: 500e9,
    board_value: 30,
    twin_value: 22,
    commands: ["cmd-1"],
    ...overrides,
  };
}

function reportSummaryWire() {
  return {
    devices: 1,
    metrics: 1,
    compared: 5,
    within: 3,
    diverged: 2,
    not_compared: 2,
    unattributed_readings: 0,
    metrics_cannot_judge: 0,
    divergences: 1,
    commands_applied: 1,
    commands_not_applied: 2,
  };
}

function reportCommandsWire() {
  return [
    { command_id: "cmd-1", device: "sensor", command: "set_mode", at_ns: 50e9, status: "applied" },
    {
      command_id: "cmd-2",
      device: "sensor",
      command: "set_mode",
      at_ns: 60e9,
      status: "not_applied",
      reason: "refused_by_board",
      detail: "not_ready",
    },
    { command_id: "cmd-3", device: "sensor", command: "reset", status: "twin_refused", detail: "unknown effector" },
  ];
}

export function twinReportWire(overrides: Record<string, unknown> = {}) {
  return {
    schema_version: "0.1",
    verdict: "diverges",
    summary: reportSummaryWire(),
    run: { field_trace_digest: digest, journal_digest: digest, evidence_bound: true },
    twin: {
      domain: "mini",
      domain_digest: digest,
      kinds: [{ kind: "thermal_sensor", kind_digest: digest }],
      start_ns: 0,
    },
    tolerance_rule: {
      default: "span_fraction",
      span_fraction: 0.05,
      description: "Absolute, per metric: 5 % of the declared range unless a tolerance is declared.",
    },
    commands: reportCommandsWire(),
    metrics: [twinMetricWire()],
    divergences: [twinDivergenceWire()],
    ...overrides,
  };
}

export function cannotJudgeWire(overrides: Record<string, unknown> = {}) {
  return twinReportWire({
    verdict: "cannot_judge",
    reason: "no_commands_applied",
    summary: { ...twinReportWire().summary, compared: 0, within: 0, diverged: 0, divergences: 0, commands_applied: 0 },
    commands: [],
    metrics: [
      twinMetricWire({
        verdict: "cannot_judge",
        summary: { compared: 0, within: 0, diverged: 0, not_compared: 3 },
        not_compared_reasons: [{ reason: "before_first_command", count: 3 }],
        series: { points: [], total: 0, truncated: false },
      }),
    ],
    divergences: [],
    ...overrides,
  });
}

export const fieldTraceReportWire = {
  ...reportWire,
  verification: {
    status: "files_match",
    checks: [
      { name: "trace", status: "match" },
      { name: "field_trace", status: "match" },
    ],
  },
};

export const sessionEntryWire = {
  type: "session",
  seq: 1,
  stream: "observer",
  at_ns: 0,
  mode: "connected",
  plan: { domain: "mini", instances: [] },
};

export const originEntryWire = {
  type: "origin",
  seq: 2,
  stream: "observer",
  at_ns: 0,
  instance: "CH-1",
  device: "sensor",
  board: "uno-07",
  environment: "hardware",
  boot_id: "61",
};

function twinSummaryModel(): TwinReport["summary"] {
  return {
    devices: 1,
    metrics: 1,
    compared: 5,
    within: 3,
    diverged: 2,
    notCompared: 2,
    unattributed: 0,
    metricsCannotJudge: 0,
    divergences: 1,
    commandsApplied: 1,
    commandsNotApplied: 2,
  };
}

function twinMetricModel(): TwinReport["metrics"][number] {
  return {
    device: "sensor",
    kind: "thermal_sensor",
    metric: "temperature",
    unit: "Cel",
    tolerance: { value: 2.5, source: "default_span" },
    verdict: "diverges",
    counts: { compared: 5, within: 3, diverged: 2, notCompared: 2 },
    notCompared: [{ reason: "before_first_command", count: 2 }],
    series: {
      points: points.map((p) => ({ atNs: p.at_ns, board: p.board, twin: p.twin })),
      total: 5,
      truncated: false,
    },
  };
}

function twinDivergenceModel(): TwinReport["divergences"][number] {
  return {
    name: "sensor.temperature#1",
    device: "sensor",
    metric: "temperature",
    firstNs: 400e9,
    lastNs: 500e9,
    readings: 2,
    tolerance: 2.5,
    worstError: 8,
    worstAtNs: 500e9,
    boardValue: 30,
    twinValue: 22,
    commands: ["cmd-1"],
  };
}

export function twinModel(overrides: Partial<TwinReport> = {}): TwinReport {
  return {
    verdict: "diverges",
    reason: null,
    summary: twinSummaryModel(),
    twinDomain: "mini",
    evidenceBound: true,
    ruleDescription: "Absolute, per metric: 5 % of the declared range unless a tolerance is declared.",
    commands: [],
    metrics: [twinMetricModel()],
    divergences: [twinDivergenceModel()],
    ...overrides,
  };
}

export function firstOf<T>(list: readonly T[]): T {
  const [first] = list;
  if (first === undefined) throw new Error("the fixture holds an entry");
  return first;
}
