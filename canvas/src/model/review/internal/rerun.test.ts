import { describe, expect, it } from "vitest";
import { blankReport } from "../../../test/runFixture";
import { rerunOf } from "./rerun";
import type { Report } from "./report";

const start = 1_767_225_600e9;

const fault = {
  kind: "fault" as const,
  action: "inject",
  injectedNs: start + 600e9,
  onsetNs: start + 600e9,
  target: "CT-1",
  name: "tower__fan_failure",
  clearedNs: null,
  clearedBy: "",
};

const setpoint = {
  atNs: start + 60e9,
  entity: "CT-1",
  effector: "set_setpoint",
  args: { tower_id: "CT-1", setpoint: 7 },
  outcome: "applied · 300 ms",
  immediate: false,
};

const refused = {
  atNs: start + 900e9,
  entity: "CH-1",
  effector: "set_mode",
  args: { chiller_id: "CH-1", mode: "eco", forced: true },
  outcome: "refused: interlock",
  immediate: false,
};

function completeRun(overrides: Partial<Report> = {}): Report {
  return blankReport({
    runId: "r-9",
    mode: "evaluate",
    ended: true,
    complete: true,
    seed: 11,
    startNs: start,
    endNs: start + 3600e9,
    requestedEndNs: start + 3600e9,
    disturbances: [fault],
    effectorCalls: [setpoint],
    ...overrides,
  });
}

describe("setting up another run from a complete run", () => {
  it("repeats the seed, mode, requested length, every fault and every effector call with its arguments", () => {
    expect(rerunOf(completeRun())).toEqual({
      fresh: false,
      left: [],
      plan: {
        seed: 11,
        durationS: 3600,
        explore: false,
        pace: "fast",
        faults: [{ entity: "CT-1", fault: "tower__fan_failure", atS: 600 }],
        calls: [{ entity: "CT-1", effector: "set_setpoint", atS: 60, args: { tower_id: "CT-1", setpoint: 7 } }],
      },
    });
  });

  it("repeats the length the run was asked for, not where an operator stopped it", () => {
    const stopped = completeRun({
      endNs: start + 1680e9,
      requestedEndNs: start + 3600e9,
      rate: 120,
      termination: { reason: "stopped_by_operator", atNs: start + 1680e9 },
      disturbances: [],
      effectorCalls: [],
    });
    const again = rerunOf(stopped);
    expect(again.plan).toMatchObject({ durationS: 3600, pace: "watch" });
    expect(again.left).toEqual(["anything planned after the stop at 00:28:00, which the run never recorded"]);
  });

  it("falls back to the last advance when the run recorded no request, and runs a paced plan at watch pace", () => {
    const unplanned = completeRun({ requestedEndNs: null, endNs: start + 1800e9, rate: 60 });
    expect(rerunOf(unplanned).plan).toMatchObject({ durationS: 1800, pace: "watch" });
    expect(rerunOf(completeRun({ rate: null })).plan.pace).toBe("fast");
  });

  it("repeats a refused call, because refusing it was part of the run", () => {
    const run = completeRun({ effectorCalls: [setpoint, refused] });
    expect(rerunOf(run).plan.calls).toEqual([
      { entity: "CT-1", effector: "set_setpoint", atS: 60, args: { tower_id: "CT-1", setpoint: 7 } },
      { entity: "CH-1", effector: "set_mode", atS: 900, args: { chiller_id: "CH-1", mode: "eco", forced: true } },
    ]);
    expect(rerunOf(run).left).toEqual([]);
  });

  it("keeps the explore mode", () => {
    expect(rerunOf(completeRun({ mode: "explore" })).plan.explore).toBe(true);
  });

  it("lists only what a launch plan cannot carry", () => {
    const run = completeRun({
      disturbances: [
        { ...fault, clearedNs: start + 900e9, clearedBy: "command" },
        { ...fault, name: "tower__other", clearedNs: start + 900e9, clearedBy: "pulse_end" },
        { ...fault, name: "tower__third", clearedNs: start + 900e9, clearedBy: "responder" },
        { ...fault, kind: "perturbation", target: "delivery", name: "drop" },
        { ...fault, kind: "environment", target: "weather", name: "heat_wave" },
      ],
    });
    expect(rerunOf(run).plan.faults).toHaveLength(3);
    expect(rerunOf(run).left).toEqual([
      "1 delivery perturbation",
      "1 environment change",
      "1 fault clear (the fault stays on)",
    ]);
  });

  it("uses one minute for a shorter run and says so", () => {
    const brief = completeRun({ requestedEndNs: start + 30e9, endNs: start + 30e9, disturbances: [] });
    expect(rerunOf(brief).plan.durationS).toBe(60);
    expect(rerunOf(brief).left).toEqual(["a length under one minute (one minute is used)"]);
  });
});

describe("setting up another run from a run that did not finish", () => {
  it("offers fresh settings in the mode the run opened with, and copies nothing of the old plan", () => {
    const interrupted = blankReport({ runId: "r-3", mode: "explore", ended: false, complete: false, startNs: start });
    expect(rerunOf(interrupted)).toEqual({
      fresh: true,
      left: [],
      plan: { seed: 1, durationS: 1800, explore: true, pace: "watch", faults: [], calls: [] },
    });
  });

  it.each([
    ["ended incomplete", { ended: true, complete: false, seed: 5, disturbances: [fault], effectorCalls: [setpoint] }],
    ["a seed the run never recorded", { ended: true, complete: true, seed: null }],
    ["no end", { ended: false, complete: true, seed: 5 }],
  ])("treats a run with %s as unrecovered and never copies its faults or calls", (_name, overrides) => {
    const rerun = rerunOf(completeRun(overrides));
    expect(rerun.fresh).toBe(true);
    expect(rerun.plan).toEqual({ seed: 1, durationS: 1800, explore: false, pace: "watch", faults: [], calls: [] });
  });
});
