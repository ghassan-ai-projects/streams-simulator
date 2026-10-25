import { describe, expect, it } from "vitest";
import { comparison, comparisonWire } from "../../../test/compareFixture";
import { parseComparisonPayload } from "./compareDecoder";

function altered(change: (wire: typeof comparisonWire) => void): unknown {
  const copy = structuredClone(comparisonWire);
  change(copy);
  return copy;
}

describe("the comparison of two runs", () => {
  it("reads the whole document the binary prints", () => {
    expect(parseComparisonPayload(comparisonWire)).toEqual(comparison);
  });

  it("reads the runs, the plan difference lane by lane and the limits in the binary's order", () => {
    const parsed = parseComparisonPayload(comparisonWire);
    expect(parsed?.a).toEqual({ runId: "w-1", name: "Baseline", startNs: 0 });
    expect(parsed?.b).toEqual({ runId: "w-2", name: "", startNs: 0 });
    expect(parsed?.plan.faults.different).toEqual([
      {
        a: { entity: "CT-1", fault: "tower__fan_failure", onsetNs: 600e9, clearedNs: null, clearedBy: "" },
        b: { entity: "CT-1", fault: "tower__fan_failure", onsetNs: 660e9, clearedNs: 900e9, clearedBy: "command" },
        differences: ["cleared_ns", "onset_ns"],
      },
    ]);
    expect(parsed?.plan.perturbations.onlyInB).toEqual([
      { action: "apply", kind: "drop", atNs: 0, fromNs: 300e9, untilNs: 600e9, params: { rate: "0.5" } },
    ]);
    expect(parsed?.plan.environment.onlyInA).toEqual([{ atNs: 900e9, target: "weather", fault: "heat_wave" }]);
    expect(parsed?.plan.effectorCalls.identical).toHaveLength(2);
    expect(parsed?.plan.effectorCalls.onlyInB[0]).toMatchObject({
      atNs: 240e9,
      entity: "CH-1",
      effector: "set_setpoint",
      args: { setpoint_c: "7" },
      outcome: "applied · 300 ms",
    });
    expect(parsed?.limits.map((limit) => [limit.code, limit.side])).toEqual([
      ["different_simulated_length", null],
      ["stopped_early", "b"],
      ["single_run_per_side", null],
    ]);
  });

  it("reads the whole-run results with the binary's own deltas and scope", () => {
    const parsed = parseComparisonPayload(comparisonWire);
    const scope = { aLengthNs: 1800e9, bLengthNs: 1680e9 };
    expect(parsed?.kpis[0]).toMatchObject({
      entity: "CH-1",
      kpi: "cop",
      presence: "both",
      delta: -0.69,
      relative: -0.1356,
      scope,
    });
    expect(parsed?.kpis[0]?.a).toEqual({ mean: 5.09, min: 4.09, max: 6.09, samples: 100 });
    expect(parsed?.kpis[1]).toMatchObject({ presence: "only_in_a", b: null, delta: null, relative: null });
    expect(parsed?.alarms[0]).toMatchObject({
      alarm: "high_temp",
      a: { priority: "high", activations: 1, activeS: 120 },
      delta: { activations: 2, activeS: 180 },
    });
    expect(parsed?.alarmSystem.delta).toMatchObject({ activations: 2, standing_at_end: 1 });
    expect(parsed?.windowedReason).toContain("whole-run KPI truth only");
  });

  it("keeps a null relative delta null instead of inventing one", () => {
    const wire = altered((copy) => {
      const row = copy.results.kpis[0];
      if (row !== undefined) row.relative_delta = { mean: null as unknown as number, min: null, max: null };
    });
    expect(parseComparisonPayload(wire)?.kpis[0]?.relative).toBeNull();
  });

  it("reads a scope whose run has no KPI truth as unknown length", () => {
    const wire = altered((copy) => {
      copy.results.alarm_system.scope = {
        kind: "whole_run",
        a_simulated_length_ns: null as unknown as number,
        b_simulated_length_ns: 5e9,
      };
    });
    expect(parseComparisonPayload(wire)?.alarmSystem.scope).toEqual({ aLengthNs: null, bLengthNs: 5e9 });
  });

  it("reads list-valued settings of a delivery condition and drops what is not a setting", () => {
    const wire = altered((copy) => {
      Object.assign(copy.plan.perturbations.only_in_b[0]?.params ?? {}, {
        fields: ["a", "b"],
        nested: { x: 1 },
        mixed: [1],
      });
    });
    expect(parseComparisonPayload(wire)?.plan.perturbations.onlyInB[0]?.params).toEqual({
      rate: "0.5",
      fields: "a, b",
    });
  });

  it("keeps an alarm system a run did not record as absent", () => {
    const wire = altered((copy) => {
      Object.assign(copy.results.alarm_system, { b: null, delta: undefined });
    });
    expect(parseComparisonPayload(wire)?.alarmSystem).toMatchObject({ b: null, delta: null });
  });

  it("ignores fields it does not know", () => {
    const wire = altered((copy) => {
      Object.assign(copy, { extra: 1 });
      Object.assign(copy.plan, { more: [] });
    });
    expect(parseComparisonPayload(wire)).not.toBeNull();
  });

  it.each([
    ["not an object", null],
    ["no plan", altered((copy) => Reflect.deleteProperty(copy, "plan"))],
    ["a plan lane that is not a difference", altered((copy) => Object.assign(copy.plan, { faults: [] }))],
    [
      "a fault item without its entity",
      altered((copy) => Object.assign(copy.plan.faults.different[0]?.a ?? {}, { entity_id: 3 })),
    ],
    ["a limit without a message", altered((copy) => Object.assign(copy.comparability_limits, { 0: { code: "x" } }))],
    [
      "a result row with an unknown presence",
      altered((copy) => Object.assign(copy.results.kpis[0] ?? {}, { presence: "neither" })),
    ],
    [
      "a result row without statistics it names",
      altered((copy) => Object.assign(copy.results.kpis[0] ?? {}, { a: { mean: "5" } })),
    ],
    [
      "a delta that is not a number",
      altered((copy) => Object.assign(copy.results.kpis[0] ?? {}, { delta: { mean: "x" } })),
    ],
    ["no windowed results", altered((copy) => Reflect.deleteProperty(copy, "windowed_results"))],
    ["no results", altered((copy) => Reflect.deleteProperty(copy, "results"))],
    [
      "a delivery condition without its action",
      altered((copy) => Object.assign(copy.plan.perturbations.only_in_b[0] ?? {}, { action: 1 })),
    ],
    [
      "an environment fault without its target",
      altered((copy) => Object.assign(copy.plan.environment.only_in_a[0] ?? {}, { target: 1 })),
    ],
    [
      "an effector call without its effector",
      altered((copy) => Object.assign(copy.plan.effector_calls.only_in_b[0] ?? {}, { effector: 1 })),
    ],
    [
      "a fault without its onset",
      altered((copy) => Object.assign(copy.plan.faults.different[0]?.a ?? {}, { onset_ns: "soon" })),
    ],
    [
      "a pair whose differences are not a list",
      altered((copy) => Object.assign(copy.plan.faults.different[0] ?? {}, { differences: "onset_ns" })),
    ],
    ["a run without its start", altered((copy) => Object.assign(copy.a, { start_time_ns: "0" }))],
    ["no limits at all", altered((copy) => Object.assign(copy, { comparability_limits: [] }))],
    [
      "an alarm row with unreadable statistics",
      altered((copy) => Object.assign(copy.results.alarms[0] ?? {}, { a: { priority: 1 } })),
    ],
    [
      "an alarm delta that is not numbers",
      altered((copy) => Object.assign(copy.results.alarms[0] ?? {}, { delta: { activations: "x" } })),
    ],
    ["an alarm system that is not counts", altered((copy) => Object.assign(copy.results.alarm_system, { a: "none" }))],
    [
      "a scope whose length is not a number",
      altered((copy) =>
        Object.assign(copy.results.kpis[0] ?? {}, { scope: { a_simulated_length_ns: "x", b_simulated_length_ns: 1 } }),
      ),
    ],
  ])("refuses a document with %s", (_name, wire) => {
    expect(parseComparisonPayload(wire)).toBeNull();
  });
});
