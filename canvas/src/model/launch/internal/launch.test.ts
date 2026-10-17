import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import {
  disturbanceWords,
  keepKnownFaults,
  launchSteps,
  outcomeOf,
  paceOf,
  planProblems,
  planSummary,
  rateWords,
  stepsFor,
  watchWords,
} from "./launch";

type PlannedCall = LaunchPlan["calls"][number];

const anySpec = (_entity: string, effector: string) => ({ name: effector, entityArg: null, args: [] });

const plan = (overrides: Partial<LaunchPlan>): LaunchPlan => ({
  seed: 7,
  durationS: 100,
  explore: true,
  faults: [],
  calls: [],
  ...overrides,
});

describe("launch plan", () => {
  it("advances in chunks to the end so progress can be shown", () => {
    const steps = launchSteps(plan({}), 4);
    expect(steps).toEqual([
      { kind: "advance", byNs: 25e9, toS: 25 },
      { kind: "advance", byNs: 25e9, toS: 50 },
      { kind: "advance", byNs: 25e9, toS: 75 },
      { kind: "advance", byNs: 25e9, toS: 100 },
    ]);
  });

  it("stops at each fault's time, injects it there, and injects time-zero faults first", () => {
    const steps = launchSteps(
      plan({
        faults: [
          { entity: "CT-1", fault: "fan", atS: 30 },
          { entity: "AC-1", fault: "cooler", atS: 0 },
        ],
      }),
      2,
    );
    expect(steps).toEqual([
      { kind: "inject", entity: "AC-1", fault: "cooler" },
      { kind: "advance", byNs: 30e9, toS: 30 },
      { kind: "inject", entity: "CT-1", fault: "fan" },
      { kind: "advance", byNs: 20e9, toS: 50 },
      { kind: "advance", byNs: 50e9, toS: 100 },
    ]);
  });

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
      { field: 0, text: "Fault 1 needs an instance and a fault." },
      { field: 1, text: "Fault 2 starts inside the run, before its end." },
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

  it("says in words what a plan will run", () => {
    expect(planSummary(plan({}), "30 min")).toBe("30 min of simulated time · Explore · no faults");
    const faults = [{ entity: "CT-1", fault: "fan", atS: 5 }];
    expect(planSummary(plan({ explore: false, faults }), "1 h")).toBe("1 h of simulated time · Evaluate · 1 fault");
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
  it("makes each call at its minute, after the faults of the same minute, at the world's clock", () => {
    const steps = launchSteps(
      plan({
        durationS: 120,
        faults: [{ entity: "CT-1", fault: "fan", atS: 60 }],
        calls: [call({ atS: 60 }), call({ atS: 0, args: { chiller_id: "CH-1", setpoint_c: 9 } })],
      }),
      1,
    );
    expect(steps).toEqual([
      { kind: "invoke", ...call({ atS: 0, args: { chiller_id: "CH-1", setpoint_c: 9 } }) },
      { kind: "advance", byNs: 60e9, toS: 60 },
      { kind: "inject", entity: "CT-1", fault: "fan" },
      { kind: "invoke", ...call({ atS: 60 }) },
      { kind: "advance", byNs: 60e9, toS: 120 },
    ]);
  });

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
      call: 0,
      part: "minute",
    });
  });

  it("says in words how many faults and effector calls a plan holds", () => {
    expect(disturbanceWords(plan({}))).toBe("no faults");
    expect(disturbanceWords(plan({ calls: [call({})] }))).toBe("no faults and 1 effector call");
    const faults = [
      { entity: "CT-1", fault: "fan", atS: 5 },
      { entity: "CT-1", fault: "fan", atS: 6 },
    ];
    expect(disturbanceWords(plan({ faults, calls: [call({}), call({})] }))).toBe("2 faults and 2 effector calls");
    expect(planSummary(plan({ calls: [call({})] }), "30 min")).toBe(
      "30 min of simulated time · Explore · no faults · 1 effector call",
    );
  });

  it("reads the acknowledgement as accepted or refused with the binary's reason", () => {
    const accepted = outcomeOf(call({}), { accepted: true, reason: "", effectEtaNs: 5e9 }, 1234);
    expect(accepted).toMatchObject({
      accepted: true,
      reason: "",
      effectEtaNs: 5e9,
      entity: "CH-1",
      atS: 60,
      atMs: 1234,
    });
    expect(outcomeOf(call({}), { accepted: false, reason: "effector_refused", effectEtaNs: null }, 0)).toMatchObject({
      accepted: false,
      reason: "effector_refused",
      effectEtaNs: null,
    });
    const silent = { accepted: false, reason: "", effectEtaNs: null };
    expect(outcomeOf(call({}), silent, 0)).toMatchObject({ accepted: false, reason: "no reason given" });
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

  it("cuts a watched run into finer steps than a fast one", () => {
    const advances = (steps: readonly { kind: string }[]) => steps.filter((step) => step.kind === "advance").length;
    expect(advances(stepsFor(plan({ durationS: 3000, pace: "watch" })))).toBe(30);
    expect(advances(stepsFor(plan({ durationS: 3000, pace: "fast" })))).toBe(20);
    expect(advances(stepsFor(plan({ durationS: 3000 })))).toBe(20);
  });
});
