import { describe, expect, it } from "vitest";
import type { CallOutcome, LaunchPlan } from "../../../test/modelFixtures";
import { showingMarks } from "../../../test/modelFixtures";
import { planMarks } from "./planMarks";

const plan: LaunchPlan = {
  seed: 1,
  durationS: 1800,
  explore: true,
  perturbations: [],
  faults: [
    { entity: "boiler-1", fault: "scaling", atS: 600 },
    { entity: "hall-a/pump-1", fault: "bearing_wear", atS: 900 },
    { entity: "", fault: "", atS: 0 },
  ],
  calls: [],
};

const call: CallOutcome = {
  entity: "pump-2",
  effector: "set_speed",
  atS: 300,
  accepted: true,
  reason: "",
  effectEtaNs: null,
  immediate: false,
  atMs: 10_000,
};

describe("plan marks", () => {
  it("marks nothing before a planned fault is due and keeps the fault from its minute to the end of the run", () => {
    expect(planMarks(showingMarks(plan, [], 599, 0)).size).toBe(0);
    for (const at of [600, 601, 1200, 1800]) {
      expect(planMarks(showingMarks(plan, [], at, 0)).get("boiler-1")).toEqual({
        status: "fault",
        count: 0,
        fault: "scaling",
        effector: "",
      });
    }
  });

  it("counts the faulted members of a hall and marks the member itself", () => {
    const marks = planMarks(showingMarks(plan, [], 950, 0));
    expect(marks.get("hall-a")).toEqual({ status: "fault", count: 1, fault: "", effector: "" });
    expect(marks.get("hall-a/pump-1")?.fault).toBe("bearing wear");
  });

  it("shows an effector call for 60 s of run time from its minute, with its outcome", () => {
    expect(planMarks(showingMarks(plan, [call], 299, 10_000)).has("pump-2")).toBe(false);
    expect(planMarks(showingMarks(plan, [call], 300, 10_000)).get("pump-2")?.effector).toBe("set speed · accepted");
    expect(planMarks(showingMarks(plan, [call], 359, 10_000)).has("pump-2")).toBe(true);
    expect(planMarks(showingMarks(plan, [call], 360, 50_000)).has("pump-2")).toBe(false);
    const refused = { ...call, accepted: false, reason: "out of range" };
    expect(planMarks(showingMarks(plan, [refused], 310, 10_000)).get("pump-2")?.effector).toBe(
      "set speed · refused: out of range",
    );
  });

  it("holds an effector call for 3 s of wall-clock time after it was made, however fast the run goes", () => {
    expect(planMarks(showingMarks(plan, [call], 480, 10_000)).has("pump-2")).toBe(true);
    expect(planMarks(showingMarks(plan, [call], 480, 12_999)).get("pump-2")?.effector).toBe("set speed · accepted");
    expect(planMarks(showingMarks(plan, [call], 480, 13_000)).has("pump-2")).toBe(false);
    expect(planMarks(showingMarks(plan, [call], 299, 11_000)).has("pump-2")).toBe(false);
  });
});

describe("plan marks in Evaluate", () => {
  const canary = "chiller__CANARY_TRUTH_9f3a";
  const blind: LaunchPlan = {
    ...plan,
    explore: false,
    faults: [{ entity: "boiler-1", fault: canary, atS: 10 }],
  };

  it("carries no fault, status or hall count while the run is unsealed, whatever the clock says", () => {
    for (const at of [0, 10, 900, 1800]) {
      const marks = planMarks(showingMarks(blind, [], at, 0));
      expect(marks.size).toBe(0);
      expect(JSON.stringify([...marks])).not.toMatch(/canary/i);
    }
  });

  it("keeps the operator's own effector call and its outcome", () => {
    const marks = planMarks(showingMarks(blind, [call], 310, 10_000));
    expect(marks.get("pump-2")).toEqual({ status: null, count: 0, fault: "", effector: "set speed · accepted" });
    expect(marks.has("boiler-1")).toBe(false);
  });

  it("still shows the same fault in Explore", () => {
    const marks = planMarks(showingMarks({ ...blind, explore: true }, [], 10, 0));
    expect(marks.get("boiler-1")?.fault).toBe("CANARY TRUTH 9f3a");
  });
});
