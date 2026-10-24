import { describe, expect, it } from "vitest";
import { failure as failed } from "../../ports";
import { cannedFailure } from "../../../test/fakeBridge";
import { deferred, flushPromises } from "../../../test/deferred";
import { launchBridge } from "../../../test/launchFixture";
import { launch, origin, plan, started } from "../../../test/launchRun";
import type { LaunchState } from "./launchState";

const explore = { ...plan, explore: true, faults: [] };

function running(state: LaunchState) {
  if (state.phase !== "running") throw new Error(`expected a running state, got ${state.phase}`);
  return state;
}

const names = (fake: ReturnType<typeof launchBridge>) => fake.calls.map((call) => call.name);

async function pausedRun(launched = explore) {
  const run = await started(launchBridge("w-4"), launched);
  await run.tick(1000);
  await run.store.pause();
  return run;
}

describe("launch store: pausing and resuming", () => {
  it("asks the binary to pause, shows the run as paused where it stands and keeps polling", async () => {
    const { store, fake, tick } = await started(launchBridge("w-4"), explore);
    await tick(1000);
    await store.pause();
    expect(fake.calls.at(-1)).toEqual({ name: "job.pause", args: { world_id: "w-4" } });
    expect(store.snapshot()).toMatchObject({ phase: "running", paused: true, atS: 60, control: { pending: null } });
    const reads = fake.jobs.statusReads;
    await tick(2000);
    expect(fake.jobs.statusReads).toBe(reads + 4);
    expect(running(store.snapshot()).atS).toBe(60);
  });

  it("goes on from where it stood when resumed, at the same rate", async () => {
    const { store, fake, tick } = await pausedRun();
    await tick(5000);
    await store.continueRun();
    expect(fake.calls.at(-1)).toEqual({ name: "job.resume", args: { world_id: "w-4" } });
    expect(store.snapshot()).toMatchObject({ paused: false, atS: 60, rate: 60 });
    await tick(1000);
    expect(running(store.snapshot()).atS).toBe(120);
  });

  it("shows pending while the call is out and takes no second control call meanwhile", async () => {
    const fake = launchBridge("w-4");
    const { store } = await started(fake, explore);
    const gate = fake.holdAction("job.pause");
    const first = store.pause();
    await flushPromises();
    expect(store.snapshot()).toMatchObject({ control: { pending: "pause", refusal: "" } });
    void store.pause();
    void store.setRate(10);
    void store.invokeNow("CH-1", "chiller__start", {});
    await flushPromises();
    gate.release();
    await first;
    expect(names(fake).filter((name) => name !== "job.status")).toEqual(["job.start", "job.pause"]);
    expect(store.snapshot()).toMatchObject({ paused: true, control: { pending: null } });
  });

  it("does not let a status read that began before a control undo the control", async () => {
    const fake = launchBridge("w-4");
    const { store, tick } = await started(fake, explore);
    const early = fake.jobs.answer("job.status", { world_id: "w-4" });
    const gate = deferred<true>();
    fake.handlers.set("job.status", () => gate.promise.then(() => early ?? failed(cannedFailure("x"))));
    await tick(500);
    fake.handlers.delete("job.status");
    await store.pause();
    gate.resolve(true);
    await flushPromises();
    expect(store.snapshot()).toMatchObject({ paused: true });
  });
});

describe("launch store: stepping", () => {
  it("steps one chunk of the run while paused and stays paused", async () => {
    const { store, fake } = await pausedRun();
    const before = running(store.snapshot()).atS;
    await store.step();
    expect(fake.calls.at(-1)).toEqual({ name: "job.step", args: { world_id: "w-4" } });
    expect(store.snapshot()).toMatchObject({ paused: true, atS: before + 60 });
  });

  it("does not ask for a step when the run is not paused", async () => {
    const { store, fake } = await started(launchBridge("w-4"), explore);
    await store.step();
    expect(names(fake)).toEqual(["job.start"]);
  });
});

describe("launch store: the speed of the run", () => {
  it("sets the live rate, 0 meaning as fast as possible, running or paused", async () => {
    const { store, fake } = await started(launchBridge("w-4"), explore);
    await store.setRate(10);
    expect(fake.calls.at(-1)).toEqual({ name: "job.rate", args: { world_id: "w-4", rate: 10 } });
    expect(store.snapshot()).toMatchObject({ rate: 10, paused: false });
    await store.pause();
    await store.setRate(300);
    expect(store.snapshot()).toMatchObject({ rate: 300, paused: true });
    await store.setRate(0);
    expect(store.snapshot()).toMatchObject({ rate: null });
  });
});

describe("launch store: a refused control", () => {
  it("keeps the director's words in the state, leaves the run as it was and clears them on the next try", async () => {
    const fake = launchBridge("w-4");
    const { store } = await started(fake, explore);
    fake.actions.set("job.pause", failed(cannedFailure("job_not_running", 409, "the job is not running or paused")));
    await store.pause();
    expect(store.snapshot()).toMatchObject({
      paused: false,
      control: { pending: null, refusal: "The simulator did not do that: the job is not running or paused." },
    });
    fake.actions.delete("job.pause");
    await store.pause();
    expect(store.snapshot()).toMatchObject({ paused: true, control: { refusal: "" } });
  });

  it("reports an unreachable bridge in the same place and does not end the run", async () => {
    const fake = launchBridge("w-4");
    const { store } = await started(fake, explore);
    fake.actions.set("job.rate", failed(cannedFailure("bridge_failed", 503, "the director is not answering")));
    await store.setRate(10);
    expect(running(store.snapshot()).control.refusal).toBe(
      "The simulator did not do that: the director is not answering.",
    );
  });
});

describe("launch store: controls end with the run", () => {
  it("takes no control once a stop is asked for", async () => {
    const { store, fake } = await started(launchBridge("w-4"), explore);
    store.stopEarly();
    await flushPromises();
    await store.pause();
    await store.setRate(10);
    expect(names(fake)).not.toContain("job.pause");
    expect(names(fake)).not.toContain("job.rate");
  });

  it("can stop a paused run, which ends where it stood", async () => {
    const { store, tick } = await pausedRun();
    store.stopEarly();
    await flushPromises();
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "done", early: true, reachedS: 60 });
  });
});

describe("launch store: commands given now", () => {
  it("calls an effector now and keeps its outcome, marked as given during the run", async () => {
    const { store, fake } = await pausedRun();
    await store.invokeNow("CH-1", "chiller__start", { chiller_id: "CH-1" });
    expect(fake.calls.at(-1)).toEqual({
      name: "job.invoke_effector",
      args: { world_id: "w-4", entity_id: "CH-1", effector: "chiller__start", args: { chiller_id: "CH-1" } },
    });
    expect(running(store.snapshot()).calls).toMatchObject([
      { entity: "CH-1", effector: "chiller__start", atS: 60, accepted: true, immediate: true },
    ]);
  });

  it("keeps a refused call as an outcome with the simulator's reason", async () => {
    const fake = launchBridge("w-4");
    fake.jobs.refuse("chiller__start", "interlock_refused");
    const { store } = await started(fake, plan);
    await store.invokeNow("CH-1", "chiller__start", {});
    expect(running(store.snapshot()).calls).toMatchObject([{ accepted: false, reason: "interlock_refused" }]);
  });

  it("injects a fault now in an Explore run and remembers it at the time it was injected", async () => {
    const { store, fake } = await pausedRun();
    await store.injectNow("CT-1", "fan_failure");
    expect(fake.calls.at(-1)).toEqual({
      name: "job.inject_fault",
      args: { world_id: "w-4", entity_id: "CT-1", fault: "fan_failure" },
    });
    expect(running(store.snapshot()).injected).toEqual([{ entity: "CT-1", fault: "fan_failure", atS: 60 }]);
  });

  it("never injects a fault in an Evaluate run, and keeps no fault to show", async () => {
    const { store, fake } = await started(launchBridge("w-4"), { ...plan, explore: false });
    await store.injectNow("CT-1", "fan_failure");
    expect(names(fake)).toEqual(["job.start"]);
    expect(running(store.snapshot()).injected).toEqual([]);
  });

  it("is lost with the tab: a run found again after a reload has no commands of this tab", async () => {
    const { store, fake } = await pausedRun();
    await store.injectNow("CT-1", "fan_failure");
    expect(running(store.snapshot()).injected).toHaveLength(1);
    const again = launch(fake);
    await again.store.resume(() => origin);
    expect(running(again.store.snapshot()).injected).toEqual([]);
  });
});

describe("launch store: finding a paused job after a reload", () => {
  const request = { domain: "factory", seed: 3, explore: true, duration_s: 1800, rate: 60, faults: [], calls: [] };

  it("takes the job over paused, at its frozen clock, and can resume it", async () => {
    const fake = launchBridge("w-4");
    const { store, tick } = launch(fake);
    fake.jobs.adopt(900, request, "paused");
    await store.resume(() => origin);
    expect(store.snapshot()).toMatchObject({ phase: "running", paused: true, atS: 900, rate: 60 });
    await tick(2000);
    expect(running(store.snapshot()).atS).toBe(900);
    await store.continueRun();
    await tick(500);
    expect(store.snapshot()).toMatchObject({ paused: false, atS: 930 });
  });

  it("stays quiet about faults in an Evaluate job found again", async () => {
    const fake = launchBridge("w-4");
    const { store } = launch(fake);
    fake.jobs.adopt(900, { ...request, explore: false }, "paused");
    await store.resume(() => origin);
    await store.injectNow("CT-1", "fan_failure");
    expect(names(fake)).not.toContain("job.inject_fault");
  });
});
