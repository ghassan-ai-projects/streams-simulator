import type { Comparison } from "../model/review";

const fault = (onset: number, cleared: number | null = null) => ({
  entity_id: "CT-1",
  fault: "tower__fan_failure",
  onset_ns: onset * 1e9,
  cleared_ns: cleared === null ? null : cleared * 1e9,
  cleared_by: cleared === null ? null : "command",
});

const call = (at: number, args: Record<string, unknown> = { setpoint_c: 7 }) => ({
  at_ns: at * 1e9,
  entity_id: "CH-1",
  effector: "set_setpoint",
  args,
  outcome: { accepted: true, effect_applied: true, ack_latency_ms: 300.4, mode: "ok" },
});

const scope = { kind: "whole_run", a_simulated_length_ns: 1800e9, b_simulated_length_ns: 1680e9 };

const round = (value: number) => Math.round(value * 1e4) / 1e4;

const stats = (mean: number) => ({ mean, min: round(mean - 1), max: round(mean + 1), samples: 100 });

const empty = { identical: [], different: [], only_in_a: [], only_in_b: [] };

export const comparisonWire = {
  schema_version: "0.1",
  a: { run_id: "w-1", name: "Baseline", mode: "explore", start_time_ns: 0, simulated_length_ns: 1800e9 },
  b: { run_id: "w-2", mode: "explore", start_time_ns: 0, simulated_length_ns: 1680e9, future: true },
  plan: {
    faults: { ...empty, different: [{ a: fault(600), b: fault(660, 900), differences: ["cleared_ns", "onset_ns"] }] },
    perturbations: {
      ...empty,
      only_in_b: [
        { action: "apply", perturbation: "drop", at_ns: 0, from_ns: 300e9, until_ns: 600e9, params: { rate: 0.5 } },
      ],
    },
    environment: {
      ...empty,
      only_in_a: [{ at_ns: 900e9, target: "weather", fault: "heat_wave" }],
    },
    effector_calls: {
      identical: [call(60), call(120)],
      different: [{ a: call(180), b: call(180, { setpoint_c: 9 }), differences: ["args"] }],
      only_in_a: [],
      only_in_b: [call(240)],
    },
  },
  comparability_limits: [
    { code: "different_simulated_length", message: "The runs simulated different lengths: A 1800 s, B 1680 s." },
    { code: "stopped_early", side: "b", message: "Run B was stopped by the operator before its requested end." },
    { code: "single_run_per_side", message: "Each side is one run with one seed; a difference may be seed noise." },
  ],
  results: {
    kpis: [
      {
        entity_id: "CH-1",
        kpi: "cop",
        presence: "both",
        a: stats(5.09),
        b: stats(4.4),
        delta: { mean: -0.69, min: -0.69, max: -0.69 },
        relative_delta: { mean: -0.1356, min: null, max: null },
        scope,
      },
      {
        entity_id: "CT-1",
        kpi: "fan_pct",
        presence: "only_in_a",
        a: stats(60),
        b: null,
        scope,
      },
      {
        entity_id: "CT-1",
        kpi: "approach_c",
        presence: "both",
        a: stats(4),
        b: stats(4.0004),
        delta: { mean: 0.0004, min: 0, max: 0 },
        relative_delta: { mean: 0.0001, min: null, max: null },
        scope,
      },
    ],
    alarms: [
      {
        entity_id: "CH-1",
        alarm: "high_temp",
        presence: "both",
        a: {
          priority: "high",
          activations: 1,
          chattering: false,
          active_s: 120,
          longest_active_s: 120,
          active_at_end: false,
        },
        b: {
          priority: "high",
          activations: 3,
          chattering: false,
          active_s: 300,
          longest_active_s: 200,
          active_at_end: true,
        },
        delta: { activations: 2, active_s: 180, longest_active_s: 80 },
        scope,
      },
    ],
    alarm_system: {
      a: { activations: 1, peak_per_10_min: 1, flood_intervals: 0, chattering: 0, stale: 0, standing_at_end: 0 },
      b: { activations: 3, peak_per_10_min: 2, flood_intervals: 0, chattering: 0, stale: 0, standing_at_end: 1 },
      delta: { activations: 2, peak_per_10_min: 1, flood_intervals: 0, chattering: 0, stale: 0, standing_at_end: 1 },
      scope,
    },
  },
  windowed_results: {
    status: "not_available",
    reason: "the run report holds whole-run KPI truth only, so no interval statistic is computed.",
  },
};

const length = { aLengthNs: 1800e9, bLengthNs: 1680e9 };
const fanFailure = { entity: "CT-1", fault: "tower__fan_failure" };
const setpoint = { atNs: 180e9, entity: "CH-1", effector: "set_setpoint", outcome: "applied · 300 ms" };

export const comparison: Comparison = {
  a: { runId: "w-1", name: "Baseline", startNs: 0 },
  b: { runId: "w-2", name: "", startNs: 0 },
  plan: {
    faults: {
      identical: [],
      different: [
        {
          a: { ...fanFailure, onsetNs: 600e9, clearedNs: null, clearedBy: "" },
          b: { ...fanFailure, onsetNs: 660e9, clearedNs: 900e9, clearedBy: "command" },
          differences: ["cleared_ns", "onset_ns"],
        },
      ],
      onlyInA: [],
      onlyInB: [],
    },
    perturbations: {
      identical: [],
      different: [],
      onlyInA: [],
      onlyInB: [{ action: "apply", kind: "drop", atNs: 0, fromNs: 300e9, untilNs: 600e9, params: { rate: "0.5" } }],
    },
    environment: {
      identical: [],
      different: [],
      onlyInA: [{ atNs: 900e9, target: "weather", fault: "heat_wave" }],
      onlyInB: [],
    },
    effectorCalls: {
      identical: [
        {
          atNs: 60e9,
          entity: "CH-1",
          effector: "set_setpoint",
          args: { setpoint_c: "7" },
          outcome: "applied · 300 ms",
        },
        {
          atNs: 120e9,
          entity: "CH-1",
          effector: "set_setpoint",
          args: { setpoint_c: "7" },
          outcome: "applied · 300 ms",
        },
      ],
      different: [
        {
          a: { ...setpoint, args: { setpoint_c: "7" } },
          b: { ...setpoint, args: { setpoint_c: "9" } },
          differences: ["args"],
        },
      ],
      onlyInA: [],
      onlyInB: [{ ...setpoint, atNs: 240e9, args: { setpoint_c: "7" } }],
    },
  },
  limits: [
    {
      code: "different_simulated_length",
      side: null,
      message: "The runs simulated different lengths: A 1800 s, B 1680 s.",
    },
    { code: "stopped_early", side: "b", message: "Run B was stopped by the operator before its requested end." },
    {
      code: "single_run_per_side",
      side: null,
      message: "Each side is one run with one seed; a difference may be seed noise.",
    },
  ],
  kpis: [
    {
      entity: "CH-1",
      kpi: "cop",
      presence: "both",
      a: { mean: 5.09, min: 4.09, max: 6.09, samples: 100 },
      b: { mean: 4.4, min: 3.4, max: 5.4, samples: 100 },
      delta: -0.69,
      relative: -0.1356,
      scope: length,
    },
    {
      entity: "CT-1",
      kpi: "fan_pct",
      presence: "only_in_a",
      a: { mean: 60, min: 59, max: 61, samples: 100 },
      b: null,
      delta: null,
      relative: null,
      scope: length,
    },
    {
      entity: "CT-1",
      kpi: "approach_c",
      presence: "both",
      a: { mean: 4, min: 3, max: 5, samples: 100 },
      b: { mean: 4.0004, min: 3.0004, max: 5.0004, samples: 100 },
      delta: 0.0004,
      relative: 0.0001,
      scope: length,
    },
  ],
  alarms: [
    {
      entity: "CH-1",
      alarm: "high_temp",
      presence: "both",
      a: { priority: "high", activations: 1, activeS: 120 },
      b: { priority: "high", activations: 3, activeS: 300 },
      delta: { activations: 2, activeS: 180 },
      scope: length,
    },
  ],
  alarmSystem: {
    a: { activations: 1, peak_per_10_min: 1, flood_intervals: 0, chattering: 0, stale: 0, standing_at_end: 0 },
    b: { activations: 3, peak_per_10_min: 2, flood_intervals: 0, chattering: 0, stale: 0, standing_at_end: 1 },
    delta: { activations: 2, peak_per_10_min: 1, flood_intervals: 0, chattering: 0, stale: 0, standing_at_end: 1 },
    scope: length,
  },
  windowedReason: "the run report holds whole-run KPI truth only, so no interval statistic is computed.",
};
