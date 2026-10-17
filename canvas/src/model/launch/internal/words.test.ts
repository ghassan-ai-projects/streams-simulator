import { describe, expect, it } from "vitest";
import type { LaunchPlan, PlanProblem } from "./launch";
import { modeWords, planCue, problemWords } from "./words";

const plan: LaunchPlan = { seed: 1, durationS: 1800, explore: true, pace: "watch", faults: [], calls: [] };

describe("launch words", () => {
  it("tells what each mode does with the true results", () => {
    expect(modeWords(true)).toBe("The truth is visible after the run; not for scoring.");
    expect(modeWords(false)).toMatch(/^The true results stay hidden from the outside system while it runs/);
  });

  it("counts the problems to fix before starting", () => {
    const problems: PlanProblem[] = [
      { field: "seed", text: "The seed is a whole number, 0 or more." },
      { field: 0, text: "Fault 1 needs a fault." },
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
    expect(planCue([duration, { field: "call", call: 1, text: "" }, { field: 2, text: "" }], plan)).toBe(
      "Duration, Effector call 2, Fault 3 need attention before starting.",
    );
  });
});
