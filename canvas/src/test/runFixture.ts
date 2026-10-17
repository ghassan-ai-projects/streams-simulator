import type { JournalEntry } from "../model/playback";
import type { Report } from "../model/report";

/** The wire report payload the bridge answers `run.report` with, as `streamsim` writes run.json. */
export const reportWire = {
  run: {
    run_id: "r-1",
    mode: "evaluate",
    complete: true,
    ended: true,
    unblinded: false,
    domain: { id: "mini" },
    adapter: { id: "native-jsonl" },
    seed: 7,
    start_time_ns: 0,
    end_time_ns: 3_600e9,
  },
  verification: { status: "files_match", checks: [{ name: "trace", status: "match" }] },
  counts: { emitted: 12, delivered: 10, dropped: 2, dropped_by_reason: { dropped_by_perturbation: 2 } },
  truth: {
    faults: [
      {
        fault_id: "f-0",
        entity_id: "CT-1",
        fault: "tower__fan_failure",
        injected_at_ns: 0,
        onset_ns: 600e9,
        cleared_ns: null,
        cleared_by: null,
      },
    ],
    perturbations: [{ action: "apply", perturbation: "drop", from_ns: 300e9 }],
    environment: [{ at_ns: 900e9, target: "weather", fault: "heat_wave" }],
    effector_calls: [
      {
        at_ns: 1_200e9,
        entity_id: "CH-1",
        effector: "start",
        outcome: { accepted: true, effect_applied: true, ack_latency_ms: 300.4, mode: "ok" },
      },
      { at_ns: 1_300e9, entity_id: "CH-1", effector: "stop", outcome: { accepted: false, refusal: "interlock" } },
      {
        at_ns: 1_400e9,
        entity_id: "CH-1",
        effector: "stop",
        outcome: { accepted: true, effect_applied: false, mode: "silent_no_effect" },
      },
      { at_ns: 1_500e9, entity_id: "CH-1", effector: "stop", outcome: { accepted: false } },
    ],
    kpis: {
      entities: [{ entity_id: "CH-1", kpis: [{ state: "cop", mean: 5.95, min: 4.65, max: 8.85 }] }],
      alarm_system: { activations: 1, standing_at_end: 0 },
    },
  },
};

/** The same report decoded at the bridge boundary; the adapter test pins the two together. */
export const report: Report = {
  runId: "r-1",
  mode: "evaluate",
  ended: true,
  complete: true,
  incompleteReason: "",
  unblinded: false,
  domainId: "mini",
  domainVersion: "",
  adapterId: "native-jsonl",
  seed: 7,
  simVersion: "",
  startNs: 0,
  endNs: 3_600e9,
  verification: "files_match",
  checks: [{ name: "trace", status: "match" }],
  counts: { emitted: 12, delivered: 10, dropped: 2 },
  dropped: { dropped_by_perturbation: 2 },
  disturbances: [
    {
      kind: "perturbation",
      action: "apply",
      injectedNs: 0,
      onsetNs: 300e9,
      target: "delivery",
      name: "drop",
      clearedNs: null,
      clearedBy: "",
    },
    {
      kind: "fault",
      action: "inject",
      injectedNs: 0,
      onsetNs: 600e9,
      target: "CT-1",
      name: "tower__fan_failure",
      clearedNs: null,
      clearedBy: "",
    },
    {
      kind: "environment",
      action: "inject",
      injectedNs: 900e9,
      onsetNs: 900e9,
      target: "weather",
      name: "heat_wave",
      clearedNs: null,
      clearedBy: "",
    },
  ],
  effectorCalls: [
    { atNs: 1_200e9, entity: "CH-1", effector: "start", outcome: "applied · 300 ms" },
    { atNs: 1_300e9, entity: "CH-1", effector: "stop", outcome: "refused: interlock" },
    { atNs: 1_400e9, entity: "CH-1", effector: "stop", outcome: "accepted, no effect · silent_no_effect" },
    { atNs: 1_500e9, entity: "CH-1", effector: "stop", outcome: "refused" },
  ],
  kpis: [{ entity: "CH-1", state: "cop", mean: 5.95, min: 4.65, max: 8.85 }],
  alarms: { activations: 1, standing_at_end: 0 },
  alarmActivations: [],
  scoring: null,
  truthUnavailable: "",
};

/** A report with every section empty, for selector tests that only care about one section. */
export function blankReport(overrides: Partial<Report> = {}): Report {
  return {
    runId: "",
    mode: "",
    ended: false,
    complete: false,
    incompleteReason: "",
    unblinded: false,
    domainId: "",
    domainVersion: "",
    adapterId: "",
    seed: null,
    simVersion: "",
    startNs: 0,
    endNs: 0,
    verification: "",
    checks: [],
    counts: {},
    dropped: {},
    disturbances: [],
    effectorCalls: [],
    kpis: [],
    alarms: {},
    alarmActivations: [],
    scoring: null,
    truthUnavailable: "",
    ...overrides,
  };
}

export function observation(seq: number, atNs: number, entity: string, channel: string, value: unknown, unit = "kW") {
  return {
    stream: "observer",
    type: "observation",
    seq,
    at_ns: atNs,
    line: JSON.stringify({ entity_id: entity, channel, value, unit }),
  };
}

/** The observer journal as the bridge streams it, line by line. */
export const journalWire = [
  { stream: "observer", type: "run.opened", seq: 0, at_ns: 0, mode: "evaluate", adapter: { id: "native-jsonl" } },
  observation(1, 10e9, "CH-1", "chiller.power", 92.7),
  observation(2, 20e9, "CH-1", "chiller.power", 95.1),
  observation(3, 20e9, "CT-1", "tower.basin_temp", 27.04, "degC"),
  observation(4, 30e9, "CT-1", "tower.run_status", true, ""),
  { stream: "observer", type: "clock.advanced", seq: 5, at_ns: 30e9 },
  { stream: "observer", type: "run.closed", seq: 6, at_ns: 3_600e9, incomplete: false },
];

/** The same journal decoded into the entry union the review store consumes. */
export const journal: readonly JournalEntry[] = [
  { kind: "opened", atNs: 0, mode: "evaluate", adapter: "native-jsonl" },
  {
    kind: "observation",
    observation: { atNs: 10e9, entity: "CH-1", channel: "chiller.power", value: 92.7, unit: "kW" },
  },
  {
    kind: "observation",
    observation: { atNs: 20e9, entity: "CH-1", channel: "chiller.power", value: 95.1, unit: "kW" },
  },
  {
    kind: "observation",
    observation: { atNs: 20e9, entity: "CT-1", channel: "tower.basin_temp", value: 27.04, unit: "degC" },
  },
  {
    kind: "observation",
    observation: { atNs: 30e9, entity: "CT-1", channel: "tower.run_status", value: true, unit: "" },
  },
  { kind: "other", atNs: 30e9 },
  { kind: "closed", atNs: 3_600e9, incomplete: false },
];
