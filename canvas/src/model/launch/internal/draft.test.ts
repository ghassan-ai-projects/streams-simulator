import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import type { Draft } from "./draft";
import { draftOf, firstDraft, pickedOnGraph, planOf } from "./draft";

const plant = parsePlant({
  id: "mini",
  instances: [{ kind: "chiller", names: ["CH-1"] }],
  kinds: [
    {
      name: "chiller",
      effectors: [
        {
          name: "chiller__set_setpoint",
          args_schema: {
            required: ["chiller_id", "setpoint_c"],
            properties: { chiller_id: { type: "string" }, setpoint_c: { type: "number", minimum: 4, maximum: 12 } },
          },
        },
      ],
    },
  ],
});

const draft = (overrides: Partial<Draft>): Draft => ({ ...firstDraft, ...overrides });

describe("launch draft", () => {
  it("turns minutes and hours typed as text into whole seconds", () => {
    expect(planOf(draft({ duration: "30" }), plant)).toMatchObject({ seed: 1, durationS: 1800, explore: true });
    expect(planOf(draft({ duration: "1.5", unit: "hours" }), plant).durationS).toBe(5400);
    expect(planOf(draft({ duration: "0.01" }), plant).durationS).toBe(1);
  });

  it("keeps an empty or unreadable number as NaN so the plan check can refuse it", () => {
    const plan = planOf(
      draft({
        seed: " ",
        duration: "",
        faults: [{ key: 1, entity: "CH-1", fault: "f", atMinutes: "x", untilMinutes: "" }],
      }),
      plant,
    );
    expect(plan.seed).toBeNaN();
    expect(plan.durationS).toBeNaN();
    expect(plan.faults[0]?.atS).toBeNaN();
  });

  it("types the arguments of a call from the effector's declaration", () => {
    const calls = [
      { key: 2, entity: "CH-1", effector: "chiller__set_setpoint", atMinutes: "5", values: { setpoint_c: " 7.5 " } },
    ];
    expect(planOf(draft({ calls, pace: "fast" }), plant)).toMatchObject({
      pace: "fast",
      calls: [
        { entity: "CH-1", effector: "chiller__set_setpoint", atS: 300, args: { chiller_id: "CH-1", setpoint_c: 7.5 } },
      ],
    });
  });

  it("fills a draft back from a plan in the unit that reads best", () => {
    const hours = draftOf({ seed: 4, durationS: 7200, explore: false, faults: [], perturbations: [], calls: [] });
    expect(hours).toMatchObject({ seed: "4", duration: "2", unit: "hours", explore: false, pace: "watch", nextKey: 1 });
    const minutes = draftOf({
      seed: 4,
      durationS: 5400,
      explore: true,
      pace: "fast",
      faults: [{ entity: "CH-1", fault: "fouling", atS: 90 }],
      perturbations: [],
      calls: [{ entity: "CH-1", effector: "chiller__set_setpoint", atS: 600, args: { setpoint_c: 7.5 } }],
    });
    expect(minutes).toMatchObject({ duration: "90", unit: "minutes", pace: "fast", nextKey: 3 });
    expect(minutes.faults).toEqual([{ key: 1, entity: "CH-1", fault: "fouling", atMinutes: "1.5", untilMinutes: "" }]);
    expect(minutes.calls).toEqual([
      { key: 2, entity: "CH-1", effector: "chiller__set_setpoint", atMinutes: "10", values: { setpoint_c: "7.5" } },
    ]);
  });

  it("gives an instance picked on the graph to the last row still waiting for one", () => {
    const faults = [
      { key: 1, entity: "CH-1", fault: "fouling", atMinutes: "0", untilMinutes: "" },
      { key: 2, entity: "", fault: "", atMinutes: "0", untilMinutes: "" },
    ];
    const calls = [{ key: 3, entity: "", effector: "", atMinutes: "0", values: {} }];
    expect(pickedOnGraph(draft({ faults, calls }), "CT-1").faults[1]).toMatchObject({ entity: "CT-1", fault: "" });
    expect(pickedOnGraph(draft({ faults: faults.slice(0, 1), calls }), "CT-1").calls[0]).toMatchObject({
      entity: "CT-1",
      effector: "",
      values: {},
    });
    const full = draft({
      faults: faults.slice(0, 1),
      calls: [{ ...calls[0], entity: "CH-1" } as Draft["calls"][number]],
    });
    expect(pickedOnGraph(full, "CT-1")).toBe(full);
    expect(pickedOnGraph(firstDraft, "CT-1")).toBe(firstDraft);
  });
});
