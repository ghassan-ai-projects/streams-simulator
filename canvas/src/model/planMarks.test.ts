import { describe, expect, it } from "vitest";
import type { CallOutcome, LaunchPlan } from "./launch";
import { planMarks } from "./planMarks";

const plan: LaunchPlan = {
  seed: 1,
  durationS: 1800,
  explore: true,
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
};

describe("plan marks", () => {
  it("marks nothing before a planned fault is due and keeps the fault from its minute to the end of the run", () => {
    expect(planMarks(plan, [], 599).size).toBe(0);
    for (const at of [600, 601, 1200, 1800]) {
      expect(planMarks(plan, [], at).get("boiler-1")).toEqual({
        status: "fault",
        count: 0,
        fault: "scaling",
        effector: "",
      });
    }
  });

  it("counts the faulted members of a hall and marks the member itself", () => {
    const marks = planMarks(plan, [], 950);
    expect(marks.get("hall-a")).toEqual({ status: "fault", count: 1, fault: "", effector: "" });
    expect(marks.get("hall-a/pump-1")?.fault).toBe("bearing wear");
  });

  it("shows an effector call for 60 s of run time from its minute, with its outcome", () => {
    expect(planMarks(plan, [call], 299).has("pump-2")).toBe(false);
    expect(planMarks(plan, [call], 300).get("pump-2")?.effector).toBe("set speed · accepted");
    expect(planMarks(plan, [call], 359).has("pump-2")).toBe(true);
    expect(planMarks(plan, [call], 360).has("pump-2")).toBe(false);
    const refused = { ...call, accepted: false, reason: "out of range" };
    expect(planMarks(plan, [refused], 310).get("pump-2")?.effector).toBe("set speed · refused: out of range");
  });
});
