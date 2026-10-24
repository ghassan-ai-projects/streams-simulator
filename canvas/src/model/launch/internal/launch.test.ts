import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { keepKnownFaults, paceOf, rateWords, watchWords } from "./launch";
import { planProblems } from "./planProblems";

type PlannedCall = LaunchPlan["calls"][number];

const anySpec = (_entity: string, effector: string) => ({ name: effector, entityArg: null, args: [] });

const plan = (overrides: Partial<LaunchPlan>): LaunchPlan => ({
  seed: 7,
  durationS: 100,
  explore: true,
  faults: [],
  perturbations: [],
  calls: [],
  ...overrides,
});

describe("launch plan", () => {
  it("names every problem of a plan in words", () => {
    expect(planProblems(plan({}), anySpec)).toEqual([]);
    const texts = (p: LaunchPlan) => planProblems(p, anySpec).map((problem) => problem.text);
    const faults = [
      { entity: "", fault: "", atS: 5 },
      { entity: "CT-1", fault: "fan", atS: 500 },
    ];
    expect(planProblems(plan({ seed: -1, durationS: 0, faults }), anySpec)).toEqual([
      { field: "seed", text: "The seed is a whole number, 0 or more." },
      { field: "duration", text: "The duration is at least one minute." },
      { field: "fault", index: 0, text: "Fault 1 needs an instance and a fault." },
      { field: "fault", index: 1, text: "Fault 2 starts inside the run, before its end." },
    ]);
    expect(texts(plan({ seed: 2 ** 53 + 1 }))).toEqual(["The seed is a whole number, 0 or more."]);
    expect(texts(plan({ durationS: 59 }))).toEqual(["The duration is at least one minute."]);
  });

  it("names only the missing part of a fault draft", () => {
    const texts = (faults: LaunchPlan["faults"]) =>
      planProblems(plan({ faults }), anySpec).map((problem) => problem.text);
    expect(texts([{ entity: "h02/M-1", fault: "", atS: 0 }])).toEqual(["Fault 1 needs a fault."]);
    expect(texts([{ entity: "", fault: "tooling_wear", atS: 0 }])).toEqual(["Fault 1 needs an instance."]);
    expect(texts([{ entity: "", fault: "", atS: 0 }])).toEqual(["Fault 1 needs an instance and a fault."]);
  });
});

const setpoint = {
  name: "chiller__set_setpoint",
  entityArg: "chiller_id",
  args: [{ name: "setpoint_c", type: "number", required: true, options: [], min: 4, max: 12 }],
} as const;

const call = (overrides: Partial<PlannedCall>): PlannedCall => ({
  entity: "CH-1",
  effector: "chiller__set_setpoint",
  atS: 60,
  args: { chiller_id: "CH-1", setpoint_c: 7 },
  ...overrides,
});

describe("effector calls in a plan", () => {
  it("names what is missing from a call and what its arguments get wrong", () => {
    const specOf = (entity: string, effector: string) =>
      entity === "CH-1" && effector === setpoint.name ? setpoint : undefined;
    const texts = (calls: PlannedCall[]) => planProblems(plan({ calls }), specOf).map((problem) => problem.text);
    expect(texts([call({})])).toEqual([]);
    expect(texts([call({ entity: "", effector: "" })])).toEqual(["Effector call 1 needs an instance and an effector."]);
    expect(texts([call({ entity: "" })])).toEqual(["Effector call 1 needs an instance."]);
    expect(texts([call({ effector: "" })])).toEqual(["Effector call 1 needs an effector."]);
    expect(texts([call({ atS: 100 })])).toEqual(["Effector call 1 is made inside the run, before its end."]);
    expect(texts([call({ effector: "chiller__stop" })])).toEqual(["Effector call 1: CH-1 has no effector stop."]);
    expect(texts([call({ args: { chiller_id: "CH-1" } })])).toEqual(["Effector call 1: setpoint c needs a value."]);
    expect(texts([call({ args: { setpoint_c: 20 } })])).toEqual([
      "Effector call 1: setpoint c is a number from 4 to 12.",
    ]);
    expect(planProblems(plan({ calls: [call({ atS: 100 })] }), anySpec)[0]).toMatchObject({
      field: "call",
      index: 0,
      part: "minute",
    });
  });
});

describe("prefilled faults", () => {
  it("keeps the faults the plant has and clears the rest, naming what was cleared", () => {
    const known = (entity: string) => (entity === "CH-1" ? ["fouling"] : null);
    const kept = keepKnownFaults(
      plan({
        faults: [
          { entity: "CH-1", fault: "fouling", atS: 10 },
          { entity: "CH-1", fault: "gone", atS: 20 },
          { entity: "PX-9", fault: "fouling", atS: 30 },
        ],
      }),
      known,
    );
    expect(kept.plan.faults).toEqual([
      { entity: "CH-1", fault: "fouling", atS: 10 },
      { entity: "CH-1", fault: "", atS: 20 },
      { entity: "", fault: "", atS: 30 },
    ]);
    expect(kept.dropped).toEqual([
      "gone on CH-1 (no longer in this plant)",
      "a fault on PX-9 (no longer in this plant)",
    ]);
  });
});

describe("run pace", () => {
  it("watches a run in about 30 s, at 60 to 3600 simulated seconds per second", () => {
    expect(paceOf(plan({ durationS: 1800, pace: "watch" }))).toBe(60);
    expect(paceOf(plan({ durationS: 86_400, pace: "watch" }))).toBe(2880);
    expect(paceOf(plan({ durationS: 600, pace: "watch" }))).toBe(60);
    expect(paceOf(plan({ durationS: 864_000, pace: "watch" }))).toBe(3600);
    expect(paceOf(plan({ durationS: 1800, pace: "fast" }))).toBeNull();
    expect(paceOf(plan({ durationS: 1800 }))).toBeNull();
  });

  it("says how long watching takes and how fast a run goes", () => {
    expect(watchWords(1800)).toBe("about 30 s");
    expect(watchWords(300)).toBe("about 5 s");
    expect(watchWords(86_400)).toBe("about 30 s");
    expect(rateWords(59.6)).toBe("60 s per s");
    expect(rateWords(null)).toBe("");
  });
});
