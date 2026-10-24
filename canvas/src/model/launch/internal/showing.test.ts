import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { shownFaults } from "./showing";

const plan = (explore: boolean): LaunchPlan => ({
  seed: 1,
  durationS: 600,
  explore,
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
