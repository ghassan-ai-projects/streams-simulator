import { describe, expect, it } from "vitest";
import type { LaunchPlan } from "./launch";
import { readinessOf, scoringWords } from "./readiness";

const plan = (overrides: Partial<LaunchPlan> = {}): LaunchPlan => ({
  seed: 7,
  durationS: 1800,
  explore: true,
  pace: "watch",
  faults: [],
  perturbations: [],
  calls: [],
  ...overrides,
});

const lines = (given: LaunchPlan) =>
  Object.fromEntries(readinessOf(given, "Factory utilities").map((l) => [l.label, l.text]));

const line = (given: LaunchPlan, label: string) => lines(given)[label];

describe("the review before a run starts", () => {
  it("names the plant, the mode and what it means for the truth, the length, the pace and the seed", () => {
    expect(lines(plan())).toEqual({
      Plant: "Factory utilities",
      Run: "Unnamed",
      Purpose: "None stated: an investigation without a stated question",
      Mode: "Explore. The truth is visible while the run is going; this run is not for scoring.",
      Length: "30 min of simulated time, seed 7",
      Pace: "Watch, about 30 s",
      Planned: "Nothing planned: the plant runs undisturbed",
      Scoring: scoringWords(true),
    });
  });

  it("counts what is planned by lane and says how many faults recover", () => {
    const planned = line(
      plan({
        faults: [
          { entity: "CT-1", fault: "a", atS: 60, untilS: 120 },
          { entity: "CT-2", fault: "b", atS: 60 },
        ],
        perturbations: [{ kind: "drop", params: {}, atS: 0 }],
        calls: [{ entity: "CH-1", effector: "start", atS: 0, args: {} }],
      }),
      "Planned",
    );
    expect(planned).toBe("2 faults (1 recovers) · 1 delivery condition · 1 effector call");
  });

  it("shows the name and purpose the operator wrote", () => {
    const named = plan({ name: "Fan failure", purpose: "Observe recovery", pace: "fast" });
    expect(line(named, "Run")).toBe("Fan failure");
    expect(line(named, "Purpose")).toBe("Observe recovery");
    expect(line(named, "Pace")).toBe("As fast as possible");
  });

  it("says that a length not yet readable is not set", () => {
    expect(line(plan({ durationS: Number.NaN, seed: Number.NaN }), "Length")).toBe("Not set");
  });

  it("states plainly that no outside system scores the run, and what to do next for an evaluation", () => {
    expect(scoringWords(true)).toBe("Unscored investigation: no outside system is scoring this run.");
    expect(line(plan({ explore: false }), "Mode")).toMatch(/^Evaluate\. /);
    expect(scoringWords(false)).toBe(
      "Unscored blind run: no outside system is scoring this run. It stays blind until it ends; to test an outside system, set up a connected session instead.",
    );
  });
});
