import { describe, expect, it } from "vitest";
import { report, reportWire } from "../../../test/runFixture";
import { parseReplayPayload, parseReportPayload } from "./reportDecoder";

function payload(entities: unknown) {
  return { run: { run_id: "r-1" }, counts: {}, verification: { status: "match" }, truth: { kpis: { entities } } };
}

describe("the run report", () => {
  it("reads the report of a run that has not ended, which carries no run id", () => {
    const open = { run: { mode: "evaluate", ended: false }, counts: {}, verification: { status: "not_checked" } };
    expect(parseReportPayload(open)).toMatchObject({ runId: "", ended: false, complete: false });
  });

  it("reads the run, counts, verification and truth sections", () => {
    const parsed = parseReportPayload(reportWire);
    expect(parsed).toEqual(report);
    if (parsed === null) throw new Error("the fixture must decode");
    expect(parsed).toMatchObject({
      runId: "r-1",
      mode: "evaluate",
      complete: true,
      domainId: "mini",
      adapterId: "native-jsonl",
      seed: 7,
      verification: "files_match",
      ended: true,
    });
    expect(parsed.counts).toEqual({ emitted: 12, delivered: 10, dropped: 2 });
    expect(parsed.dropped).toEqual({ dropped_by_perturbation: 2 });
    expect(
      parsed.disturbances.map((d) => [d.onsetNs / 1e9, d.injectedNs / 1e9, d.kind, d.action, d.target, d.name]),
    ).toEqual([
      [300, 0, "perturbation", "apply", "delivery", "drop"],
      [600, 0, "fault", "inject", "CT-1", "tower__fan_failure"],
      [900, 900, "environment", "inject", "weather", "heat_wave"],
    ]);
    expect(parsed.effectorCalls.map((call) => call.outcome)).toEqual([
      "applied · 300 ms",
      "refused: interlock",
      "accepted, no effect · silent_no_effect",
      "refused",
    ]);
    expect(parsed.kpis).toEqual([{ entity: "CH-1", state: "cop", mean: 5.95, min: 4.65, max: 8.85 }]);
    expect(parsed.alarms).toEqual({ activations: 1, standing_at_end: 0 });
    expect(parsed.scoring).toBeNull();
  });

  it("reads the plan and the ending a job-run recorded, and nothing for a run that recorded none", () => {
    const stopped = {
      ...reportWire,
      run: {
        ...reportWire.run,
        requested_end_ns: 7_200e9,
        termination: { reason: "stopped_by_operator", at_ns: 1_680e9 },
      },
    };
    expect(parseReportPayload(stopped)).toMatchObject({
      requestedEndNs: 7_200e9,
      termination: { reason: "stopped_by_operator", atNs: 1_680e9 },
    });
    expect(parseReportPayload(reportWire)).toMatchObject({ requestedEndNs: null, termination: null });
  });

  it("reads the pace a paced run recorded, and none for a run that ran as fast as it could", () => {
    const paced = { ...reportWire, run: { ...reportWire.run, rate: 60 } };
    expect(parseReportPayload(paced)).toMatchObject({ rate: 60 });
    expect(parseReportPayload(reportWire)).toMatchObject({ rate: null });
    expect(parseReportPayload({ ...reportWire, run: { ...reportWire.run, rate: "fast" } })).toMatchObject({
      rate: null,
    });
  });

  it("reads each effector call's arguments, keeping only the numbers, texts and booleans a plan can carry", () => {
    const call = (args: unknown) => ({
      at_ns: 60e9,
      entity_id: "CT-1",
      effector: "set_setpoint",
      args,
      outcome: { accepted: true },
    });
    const wire = (args: unknown) => ({ ...reportWire, truth: { ...reportWire.truth, effector_calls: [call(args)] } });
    const argsOf = (args: unknown) => parseReportPayload(wire(args))?.effectorCalls[0]?.args;
    expect(argsOf({ tower_id: "CT-1", setpoint: 7, eco: true, mode: "auto" })).toEqual({
      tower_id: "CT-1",
      setpoint: 7,
      eco: true,
      mode: "auto",
    });
    expect(argsOf({ nested: { a: 1 }, list: [1], none: null, kept: 2 })).toEqual({ kept: 2 });
    expect(argsOf(undefined)).toEqual({});
    expect(argsOf("7")).toEqual({});
  });

  it.each([
    ["an unknown reason", { reason: "abandoned", at_ns: 5 }],
    ["a text instant", { reason: "completed", at_ns: "5" }],
    ["no instant", { reason: "completed" }],
    ["a text object", "completed"],
  ])("ignores a termination with %s", (_name, termination) => {
    const wire = { ...reportWire, run: { ...reportWire.run, requested_end_ns: "soon", termination } };
    expect(parseReportPayload(wire)).toMatchObject({ requestedEndNs: null, termination: null });
  });

  it("refuses a report without the sections a reader needs", () => {
    expect(parseReportPayload(null)).toBeNull();
    expect(parseReportPayload({})).toBeNull();
    expect(parseReportPayload({ run: { run_id: "r-1" }, counts: {} })).toBeNull();
    expect(parseReportPayload({ counts: {}, verification: { status: "match" } })).toBeNull();
    expect(
      parseReportPayload({ run: { run_id: "r-1" }, counts: "none", verification: { status: "match" } }),
    ).toBeNull();
  });

  it("keeps what a scored or truthless report says", () => {
    const base = { run: { run_id: "r-1" }, counts: {}, verification: { status: "" } };
    const scored = parseReportPayload({
      ...base,
      scoring: { status: "scored", reason: "verdict r-1" },
      truth_unavailable: "sealed",
    });
    expect(scored?.scoring).toEqual({ status: "scored", reason: "verdict r-1" });
    expect(scored?.truthUnavailable).toBe("sealed");
    expect(parseReportPayload({ ...base, scoring: { status: "failed" } })?.scoring).toEqual({
      status: "failed",
      reason: "",
    });
    expect(parseReportPayload(base)?.disturbances).toEqual([]);
  });

  it("keeps the binary's alarm instants, entity and alarm identity", () => {
    const alarm = { state: "temp_alarm", priority: "high", activation_times_ns: [140e9, 530e9] };
    expect(parseReportPayload(payload([{ entity_id: "CH-1", alarms: [alarm] }]))?.alarmActivations).toEqual([
      { atNs: 140e9, entity: "CH-1", state: "temp_alarm", priority: "high" },
      { atNs: 530e9, entity: "CH-1", state: "temp_alarm", priority: "high" },
    ]);
  });

  it("does not invent alarm instants from counts or invalid values", () => {
    expect(parseReportPayload(payload(null))?.alarmActivations).toEqual([]);
    expect(parseReportPayload(payload([null, { alarms: [null, { activations: 3 }] }]))?.alarmActivations).toEqual([]);
    expect(
      parseReportPayload(payload([{ alarms: [{ activation_times_ns: ["1", null, Infinity, 2] }] }]))?.alarmActivations,
    ).toEqual([{ atNs: 2, entity: "", state: "", priority: "" }]);
  });
});

describe("the replay verdict", () => {
  it("maps the binary's verdict, or refuses a payload without one", () => {
    expect(parseReplayPayload({ matches: true, reason: "match" })).toEqual({
      matches: true,
      domainChanged: false,
      detail: "",
      reason: "match",
      file: "",
    });
    expect(
      parseReplayPayload({ matches: false, domain_changed: true, detail: "trace differs", file: "trace.jsonl" }),
    ).toEqual({
      matches: false,
      domainChanged: true,
      detail: "trace differs",
      reason: "",
      file: "trace.jsonl",
    });
    expect(parseReplayPayload({ matches: "yes" })).toBeNull();
    expect(parseReplayPayload(null)).toBeNull();
  });
});
