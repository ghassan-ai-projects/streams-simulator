export const report = {
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

export function observation(seq: number, atNs: number, entity: string, channel: string, value: unknown, unit = "kW") {
  return {
    type: "observation",
    seq,
    at_ns: atNs,
    line: JSON.stringify({ entity_id: entity, channel, value, unit }),
  };
}

export const journal = [
  { type: "run.opened", seq: 0, at_ns: 0, mode: "evaluate", adapter: { id: "native-jsonl" } },
  observation(1, 10e9, "CH-1", "chiller.power", 92.7),
  observation(2, 20e9, "CH-1", "chiller.power", 95.1),
  observation(3, 20e9, "CT-1", "tower.basin_temp", 27.04, "degC"),
  observation(4, 30e9, "CT-1", "tower.run_status", true, ""),
  { type: "clock.advanced", seq: 5, at_ns: 30e9 },
  { type: "run.closed", seq: 6, at_ns: 3_600e9, incomplete: false },
];
