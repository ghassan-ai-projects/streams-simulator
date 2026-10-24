import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { progressTicks } from "./ticks";

const plan: LaunchPlan = {
  seed: 1,
  durationS: 600,
  explore: true,
  faults: [{ entity: "CH-1", fault: "chiller__CANARY_TRUTH_9f3a", atS: 120 }],
  perturbations: [],
  calls: [{ entity: "CH-1", effector: "start", atS: 60, args: {} }],
};

describe("progress ticks", () => {
  it("lists the planned fault and the planned call in Explore", () => {
    expect(progressTicks(plan)).toEqual([
      { atS: 120, call: false, title: "CH-1 CANARY TRUTH 9f3a at 00:02:00" },
      { atS: 60, call: true, title: "CH-1 start at 00:01:00" },
    ]);
  });

  it("lists only the operator's own calls in Evaluate, so no fault name, entity or minute reaches a tooltip", () => {
    const ticks = progressTicks({ ...plan, explore: false });
    expect(ticks).toEqual([{ atS: 60, call: true, title: "CH-1 start at 00:01:00" }]);
    expect(JSON.stringify(ticks)).not.toMatch(/canary/i);
  });
});
