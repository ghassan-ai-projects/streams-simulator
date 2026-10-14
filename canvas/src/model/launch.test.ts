import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { clockOf, launchSteps, planProblems } from "./launch";

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
    expect(planProblems(plan({ seed: -1, durationS: 0, faults: [{ entity: "", fault: "", atS: 5 }] }))).toEqual([
      "The seed is a whole number, 0 or more.",
      "The duration is at least one second.",
      "Fault 1 needs an instance and a fault.",
      "Fault 1 starts inside the run, before its end.",
    ]);
  });
});

describe("launch clock", () => {
  it("prints simulated seconds as hours, minutes and seconds", () => {
    expect(clockOf(0)).toBe("00:00:00");
    expect(clockOf(3725)).toBe("01:02:05");
    expect(clockOf(-3)).toBe("00:00:00");
  });
});
