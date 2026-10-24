import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { shownFaults, showingMarks } from "./showing";

const plan = (explore: boolean): LaunchPlan => ({
  seed: 1,
  durationS: 600,
  explore,
  perturbations: [],
  faults: [
    { entity: "CT-1", fault: "fan_failure", atS: 240 },
    { entity: "", fault: "fan_failure", atS: 10 },
    { entity: "CH-1", fault: "", atS: 20 },
  ],
  calls: [],
});

describe("the planned faults a surface may show", () => {
  it("are those of an Explore plan that name an instance and a fault", () => {
    expect(shownFaults(plan(true))).toEqual([{ entity: "CT-1", fault: "fan_failure", atS: 240 }]);
  });

  it("are none for an Evaluate plan, whose faults stay sealed", () => {
    expect(shownFaults(plan(false))).toEqual([]);
  });
});

describe("the faults injected during the run", () => {
  const injected = [{ entity: "CH-1", fault: "fouling", atS: 700 }];

  it("join the planned ones of an Explore run, and show on their instance from their time", () => {
    expect(shownFaults(plan(true), injected)).toEqual([
      { entity: "CT-1", fault: "fan_failure", atS: 240 },
      ...injected,
    ]);
    expect(showingMarks(plan(true), [], 699, 0, injected).faults.has("CH-1")).toBe(false);
    expect(showingMarks(plan(true), [], 700, 0, injected).faults.get("CH-1")).toEqual(["fouling"]);
  });

  it("are never shown for an Evaluate run, whatever was passed", () => {
    expect(shownFaults(plan(false), injected)).toEqual([]);
    expect(showingMarks(plan(false), [], 900, 0, injected).faults.size).toBe(0);
  });
});
