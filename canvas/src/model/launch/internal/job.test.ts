import { describe, expect, it } from "vitest";
import type { JobStatus } from "./job";
import {
  jobActive,
  jobFailureWords,
  jobRequestOf,
  observedCalls,
  pacedRate,
  plannedS,
  reachedS,
  startFailureWords,
} from "./job";
import type { LaunchPlan } from "./launch";
import { planSummary } from "./words";

const plan: LaunchPlan = {
  seed: 3,
  durationS: 1800,
  explore: false,
  pace: "watch",
  faults: [{ entity: "CT-1", fault: "fan", atS: 10 }],
  calls: [{ entity: "CH-1", effector: "chiller__start", atS: 20, args: { chiller_id: "CH-1" } }],
};

const status: JobStatus = {
  worldId: "w-1",
  runId: "r-1",
  domain: "factory",
  seed: 3,
  state: "running",
  startNs: 100e9,
  reachedNs: 700e9,
  requestedEndNs: 1900e9,
  rate: 60,
  explore: false,
  calls: [],
  failure: "",
};

describe("job request", () => {
  it("carries the plan and the pace as a rate, and asks for no pace as rate 0", () => {
    expect(jobRequestOf("factory", plan)).toEqual({
      domain: "factory",
      seed: 3,
      explore: false,
      durationS: 1800,
      rate: 60,
      faults: plan.faults,
      calls: plan.calls,
    });
    expect(jobRequestOf("factory", { ...plan, pace: "fast" }).rate).toBe(0);
    expect(jobRequestOf("factory", { seed: 3, durationS: 1800, explore: false, faults: [], calls: [] }).rate).toBe(0);
  });
});

describe("job status", () => {
  it("measures simulated time from the job's own start", () => {
    expect(reachedS(status)).toBe(600);
    expect(plannedS(status)).toBe(1800);
  });

  it("is active while running or stopping, and paced when the rate is positive", () => {
    expect(jobActive(status)).toBe(true);
    expect(jobActive({ ...status, state: "stopping" })).toBe(true);
    expect(jobActive({ ...status, state: "completed" })).toBe(false);
    expect(jobActive({ ...status, state: "stopped" })).toBe(false);
    expect(jobActive({ ...status, state: "failed" })).toBe(false);
    expect(pacedRate(status)).toBe(60);
    expect(pacedRate({ ...status, rate: 0 })).toBeNull();
  });
});

describe("call outcomes", () => {
  const accepted = {
    entity: "CH-1",
    effector: "chiller__start",
    atS: 20,
    accepted: true,
    reason: "",
    effectEtaNs: 5e9,
  };
  const refused = { ...accepted, accepted: false, reason: "interlock_refused", effectEtaNs: null };

  it("stamps a new outcome with the wall time it was first seen and keeps the earlier ones as they were", () => {
    const first = observedCalls([], { ...status, calls: [accepted] }, 1000);
    expect(first).toEqual([{ ...accepted, atMs: 1000 }]);
    const second = observedCalls(first, { ...status, calls: [accepted, refused] }, 2500);
    expect(second).toEqual([
      { ...accepted, atMs: 1000 },
      { ...refused, atMs: 2500 },
    ]);
    expect(observedCalls(second, { ...status, calls: [accepted, refused] }, 9000)).toBe(second);
  });

  it("gives a refusal without a reason a reason", () => {
    const silent = { ...refused, reason: "" };
    expect(observedCalls([], { ...status, calls: [silent] }, 0)[0]?.reason).toBe("no reason given");
  });
});

describe("failure words", () => {
  it("quotes the director's reason once, ending with a full stop", () => {
    expect(startFailureWords("job: duration_s must be a positive number of seconds")).toBe(
      "job: duration_s must be a positive number of seconds. Check the plan and start again.",
    );
    expect(jobFailureWords("job failed: boom. ")).toBe("job failed: boom. Check the plan and start again.");
    expect(jobFailureWords(" ")).toBe("The job stopped without saying why. Check the plan and start again.");
  });
});

describe("a plan recovered after a reload", () => {
  it("names the length and the mode but claims no fault count it does not know", () => {
    expect(planSummary(plan, "30 min")).toBe("30 min of simulated time · Evaluate · 1 fault · 1 effector call");
    expect(planSummary({ ...plan, recovered: true, faults: [], calls: [] }, "30 min")).toBe(
      "30 min of simulated time · Evaluate",
    );
  });
});
