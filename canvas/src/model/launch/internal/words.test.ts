import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import type { PlanProblem } from "./planProblems";
import { disturbanceWords, modeWords, planCue, planSummary, problemWords } from "./words";

const plan: LaunchPlan = { seed: 1, durationS: 1800, explore: true, pace: "watch", faults: [], calls: [] };

const faulted = (count: number): LaunchPlan["faults"] =>
  Array.from({ length: count }, (_, index) => ({ entity: "CT-1", fault: "fan", atS: 5 + index }));

const called = (count: number): LaunchPlan["calls"] =>
  Array.from({ length: count }, () => ({ entity: "CH-1", effector: "chiller__set_setpoint", atS: 60, args: {} }));

describe("launch words", () => {
  it("tells what each mode does with the true results", () => {
    expect(modeWords(true)).toBe("The truth is visible while the run is going; this run is not for scoring.");
    expect(modeWords(false)).toMatch(/^The true results stay hidden from the outside system while it runs/);
  });

  it("counts the problems to fix before starting", () => {
    const problems: PlanProblem[] = [
      { field: "seed", text: "The seed is a whole number, 0 or more." },
      { field: "fault", index: 0, text: "Fault 1 needs a fault." },
    ];
    expect(problemWords([])).toBe("");
    expect(problemWords(problems)).toBe(
      "2 to fix before starting: The seed is a whole number, 0 or more. Fault 1 needs a fault.",
    );
  });

  it("describes what Start will do, or names the fields that need attention", () => {
    expect(planCue([], plan)).toBe(
      "Start simulates 30 min of plant time in Explore mode with no faults, paced to take about 30 s, then saves the run and opens its review.",
    );
    expect(planCue([], { ...plan, durationS: 0, explore: false, pace: "fast" })).toBe(
      "Start simulates the chosen time of plant time in Evaluate mode with no faults, as fast as it can, then saves the run and opens its review.",
    );
    const seed: PlanProblem = { field: "seed", text: "" };
    const duration: PlanProblem = { field: "duration", text: "" };
    expect(planCue([seed], plan)).toBe("Seed needs attention before starting.");
    expect(
      planCue(
        [duration, { field: "call", index: 1, part: "args", text: "" }, { field: "fault", index: 2, text: "" }],
        plan,
      ),
    ).toBe("Duration, Effector call 2, Fault 3 need attention before starting.");
  });

  it("says in words what a plan will run", () => {
    expect(planSummary(plan, "30 min")).toBe("30 min of simulated time · Explore · no faults");
    const evaluated = { ...plan, explore: false, faults: faulted(1) };
    expect(planSummary(evaluated, "1 h")).toBe("1 h of simulated time · Evaluate · 1 fault");
  });

  it("says in words how many faults and effector calls a plan holds", () => {
    expect(disturbanceWords(plan)).toBe("no faults");
    expect(disturbanceWords({ ...plan, calls: called(1) })).toBe("no faults and 1 effector call");
    expect(disturbanceWords({ ...plan, faults: faulted(2), calls: called(2) })).toBe("2 faults and 2 effector calls");
    expect(planSummary({ ...plan, calls: called(1) }, "30 min")).toBe(
      "30 min of simulated time · Explore · no faults · 1 effector call",
    );
  });
});
