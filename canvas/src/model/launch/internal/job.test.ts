import { describe, expect, it } from "vitest";
import type { JobStatus } from "./job";
import {
  jobActive,
  jobFailureWords,
  jobPaused,
  jobRequestOf,
  observedCalls,
  observedInjected,
  pacedRate,
  plannedS,
  reachedS,
  recoveredPlanOf,
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
  perturbations: [],
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
  name: "",
  purpose: "",
  faults: [],
  perturbations: [],
  plannedCalls: [],
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
      perturbations: [],
      calls: plan.calls,
    });
    expect(jobRequestOf("factory", { ...plan, pace: "fast" }).rate).toBe(0);
    expect(
      jobRequestOf("factory", { seed: 3, durationS: 1800, explore: false, faults: [], perturbations: [], calls: [] })
        .rate,
    ).toBe(0);
  });

  it("carries the name, the purpose, the end of a fault and the delivery conditions when the plan has them", () => {
    const full: LaunchPlan = {
      ...plan,
      name: "Fan failure",
      purpose: "Observe recovery",
      faults: [{ entity: "CT-1", fault: "fan", atS: 10, untilS: 90 }],
      perturbations: [{ kind: "drop", params: { rate: 0.5 }, atS: 30, untilS: 120 }],
    };
    expect(jobRequestOf("factory", full)).toMatchObject({
      name: "Fan failure",
      purpose: "Observe recovery",
      faults: [{ entity: "CT-1", fault: "fan", atS: 10, untilS: 90 }],
      perturbations: [{ kind: "drop", params: { rate: 0.5 }, atS: 30, untilS: 120 }],
    });
    const bare = jobRequestOf("factory", plan);
    expect("name" in bare).toBe(false);
    expect("purpose" in bare).toBe(false);
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
    expect(jobActive({ ...status, state: "paused" })).toBe(true);
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
    immediate: false,
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

describe("a paused job and commands given now", () => {
  it("is paused only in the paused state", () => {
    expect(jobPaused({ ...status, state: "paused" })).toBe(true);
    expect(jobPaused(status)).toBe(false);
  });

  it("keeps the mark of a command given during the run on its outcome", () => {
    const now = { entity: "CH-1", effector: "chiller__stop", atS: 300, accepted: true, reason: "", effectEtaNs: 1e9 };
    const seen = observedCalls([], { ...status, calls: [{ ...now, immediate: true }] }, 7);
    expect(seen).toEqual([{ ...now, immediate: true, atMs: 7 }]);
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
  const sealed = { ...status, name: "Fan failure", purpose: "Observe recovery" };

  it("keeps the name and purpose and claims nothing of an evaluation's sealed plan", () => {
    expect(recoveredPlanOf(sealed)).toEqual({
      name: "Fan failure",
      purpose: "Observe recovery",
      seed: 3,
      durationS: 1800,
      explore: false,
      faults: [],
      perturbations: [],
      calls: [],
      recovered: true,
    });
  });

  it("rebuilds the faults and delivery conditions of an exploration from its status", () => {
    const open = {
      ...sealed,
      explore: true,
      faults: [{ entity: "CT-1", fault: "fan", atS: 10, untilS: 90 }],
      perturbations: [{ kind: "drop", params: { rate: 0.5 }, atS: 30 }],
    };
    expect(recoveredPlanOf(open)).toMatchObject({
      faults: open.faults,
      perturbations: open.perturbations,
      recovered: true,
    });
    expect(planSummary(recoveredPlanOf(open), "30 min")).toBe(
      "30 min of simulated time · Explore · 1 fault · 1 delivery condition",
    );
  });

  it("brings back the operator's planned calls in every mode, and leaves the live faults out of the plan", () => {
    const calls = [{ entity: "CH-1", effector: "set_setpoint", atS: 60, args: { setpoint_c: 7 } }];
    const evaluation = recoveredPlanOf({ ...status, plannedCalls: calls });
    expect(evaluation).toMatchObject({ calls, faults: [], perturbations: [] });
    expect(planSummary(evaluation, "30 min")).toBe("30 min of simulated time · Evaluate · 1 effector call");
    const open = {
      ...status,
      explore: true,
      plannedCalls: calls,
      faults: [
        { entity: "CT-1", fault: "fan", atS: 120, untilS: 540, clearedAtS: 540 },
        { entity: "CH-1", fault: "fouling", atS: 600, immediate: true, faultId: "f-1" },
      ],
    };
    expect(recoveredPlanOf(open).faults).toEqual([open.faults[0]]);
    expect(planSummary(recoveredPlanOf(open), "30 min")).toBe(
      "30 min of simulated time · Explore · 1 fault · 1 effector call",
    );
  });

  it("names the length and the mode but claims no fault count it does not know", () => {
    expect(planSummary(plan, "30 min")).toBe("30 min of simulated time · Evaluate · 1 fault · 1 effector call");
    expect(planSummary({ ...plan, recovered: true, faults: [], perturbations: [], calls: [] }, "30 min")).toBe(
      "30 min of simulated time · Evaluate",
    );
  });
});

describe("the faults injected now, as the status lists them", () => {
  const live = { entity: "CH-1", fault: "fouling", atS: 600, immediate: true, faultId: "f-1" };
  const open: JobStatus = {
    ...status,
    explore: true,
    faults: [{ entity: "CT-1", fault: "fan", atS: 120 }, live],
  };

  it("are those flagged immediate, with the second they were injected and their id", () => {
    expect(observedInjected([], open)).toEqual([{ entity: "CH-1", fault: "fouling", atS: 600, faultId: "f-1" }]);
  });

  it("are not added again when the same id, or the same instance, fault and second, is already known", () => {
    const known = [{ entity: "CH-1", fault: "fouling", atS: 600.4, faultId: "f-1" }];
    expect(observedInjected(known, open)).toBe(known);
    const unnamed = [{ entity: "CH-1", fault: "fouling", atS: 600.4 }];
    const anonymous = { entity: "CH-1", fault: "fouling", atS: 600, immediate: true };
    expect(observedInjected(unnamed, { ...open, faults: [anonymous] })).toBe(unnamed);
  });

  it("add a second injection of the same fault under its own id", () => {
    const second = { ...live, atS: 700, faultId: "f-2" };
    const known = observedInjected([], open);
    expect(observedInjected(known, { ...open, faults: [live, second] })).toHaveLength(2);
  });

  it("are none while an Evaluate job hides its faults", () => {
    expect(observedInjected([], { ...status, faults: [] })).toEqual([]);
  });
});
