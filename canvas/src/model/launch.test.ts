import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { clockOf, launchSteps, planProblems, planSummary } from "./launch";

const plan = (overrides: Partial<LaunchPlan>): LaunchPlan => ({
  seed: 7,
  durationS: 100,
  explore: true,
  faults: [],
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
    expect(planProblems(plan({}))).toEqual([]);
    const texts = (p: LaunchPlan) => planProblems(p).map((problem) => problem.text);
    const faults = [
      { entity: "", fault: "", atS: 5 },
      { entity: "CT-1", fault: "fan", atS: 500 },
    ];
    expect(planProblems(plan({ seed: -1, durationS: 0, faults }))).toEqual([
      { field: "seed", text: "The seed is a whole number, 0 or more." },
      { field: "duration", text: "The duration is at least one minute." },
      { field: 0, text: "Fault 1 needs an instance and a fault." },
      { field: 1, text: "Fault 2 starts inside the run, before its end." },
    ]);
    expect(texts(plan({ seed: 2 ** 53 + 1 }))).toEqual(["The seed is a whole number, 0 or more."]);
    expect(texts(plan({ durationS: 59 }))).toEqual(["The duration is at least one minute."]);
  });

  it("names only the missing part of a fault draft", () => {
    const texts = (faults: LaunchPlan["faults"]) => planProblems(plan({ faults })).map((problem) => problem.text);
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

describe("launch clock", () => {
  it("prints simulated seconds as hours, minutes and seconds", () => {
    expect(clockOf(0)).toBe("00:00:00");
    expect(clockOf(3725)).toBe("01:02:05");
    expect(clockOf(-3)).toBe("00:00:00");
  });
});
