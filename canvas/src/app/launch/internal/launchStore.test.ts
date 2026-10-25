import { describe, expect, it } from "vitest";
import { failure as failed } from "../../ports";
import { cannedFailure } from "../../../test/fakeBridge";
import { flushPromises } from "../../../test/deferred";
import { launchBridge } from "../../../test/launchFixture";
import { launch, origin, plan, start, started } from "../../../test/launchRun";

describe("launch store: starting a job", () => {
  it("asks the binary for the whole plan at once, paced as the watch pace says, and does no stepping itself", async () => {
    const { fake, store } = await started();
    expect(fake.calls.map((call) => call.name)).toEqual(["job.start"]);
    expect(fake.calls[0]?.args).toEqual({
      domain: "factory",
      seed: 3,
      explore: false,
      duration_s: 1800,
      rate: 60,
      faults: [{ entity_id: "CT-1", fault: "fan", at_s: 600 }],
      perturbations: [],
      calls: [],
    });
    expect(store.snapshot()).toMatchObject({ phase: "running", worldId: "w-4", atS: 0, durationS: 1800, rate: 60 });
  });

  it("shows the run as starting, without a world, until the binary answers", async () => {
    const fake = launchBridge("w-4");
    const gate = fake.holdAction("job.start");
    const { store } = launch(fake);
    const starting = store.start(origin, plan);
    expect(store.snapshot()).toMatchObject({ phase: "running", worldId: null, atS: 0, stopping: false, calls: [] });
    gate.release();
    await starting;
    expect(store.snapshot()).toMatchObject({ phase: "running", worldId: "w-4" });
  });

  it("asks for rate 0, as fast as it can, when the plan has no pace", async () => {
    const { fake, store } = await started(launchBridge("w-4"), { ...plan, pace: "fast" });
    expect(fake.calls[0]?.args.rate).toBe(0);
    expect(store.snapshot()).toMatchObject({ phase: "running", rate: null });
  });

  it("ignores a second start while a run is going", async () => {
    const { fake, store } = await started();
    await store.start(origin, plan);
    expect(fake.calls.filter((call) => call.name === "job.start")).toHaveLength(1);
  });

  it("fails without a world, in the director's words, when the job is refused", async () => {
    const fake = launchBridge("w-4");
    fake.actions.set("job.start", failed(cannedFailure("job_active", 409, "a job is already running or stopping")));
    const { store, timers } = launch(fake);
    await store.start(origin, plan);
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      origin,
      failure: { kind: "rejected", step: "start", diagnostic: { displayCode: "job_active" } },
      words: "a job is already running or stopping. Check the plan and start again.",
      kept: null,
    });
    expect(timers.pending()).toBe(0);
  });
});

describe("launch store: following the job", () => {
  it("polls the status twice a second and shows the simulated time the binary has reached", async () => {
    const { fake, store, tick } = await started();
    await tick(500);
    expect(fake.jobs.statusReads).toBe(1);
    expect(store.snapshot()).toMatchObject({ phase: "running", atS: 30, durationS: 1800, rate: 60, startedMs: 0 });
    await tick(500);
    expect(fake.jobs.statusReads).toBe(2);
    expect(store.snapshot()).toMatchObject({ atS: 60, atMs: 1000 });
    expect(fake.calls.find((call) => call.name === "job.status")?.args).toEqual({ world_id: "w-4" });
  });

  it("finishes with the run's own numbers when the job completes, and stops polling", async () => {
    const { fake, store, tick, timers } = await started();
    await tick(30_000);
    expect(store.snapshot()).toEqual({
      phase: "done",
      origin,
      worldId: "w-4",
      early: false,
      plan,
      reachedS: 1800,
      calls: [],
    });
    const reads = fake.jobs.statusReads;
    await tick(10_000);
    expect(fake.jobs.statusReads).toBe(reads);
    expect(timers.pending()).toBe(0);
  });

  it("finishes at once when a run as fast as it can is already over at the first poll", async () => {
    const { store, tick } = await started(launchBridge("w-4"), { ...plan, pace: "fast" });
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "done", early: false, reachedS: 1800 });
  });

  it("keeps a refused effector call as an outcome, stamped with the wall time it was first seen", async () => {
    const fake = launchBridge("w-4");
    fake.jobs.refuse("chiller__stop", "interlock_refused");
    const stop = { ...start, effector: "chiller__stop", atS: 900 };
    const { store, tick } = await started(fake, { ...plan, faults: [], calls: [stop, start] });
    await tick(5000);
    expect(store.snapshot()).toMatchObject({ phase: "running", calls: [] });
    await tick(5000);
    const running = store.snapshot();
    expect(running.phase === "running" ? running.calls : []).toEqual([
      {
        entity: "CH-1",
        effector: "chiller__start",
        atS: 600,
        accepted: true,
        reason: "",
        effectEtaNs: 30e9,
        immediate: false,
        atMs: 10_000,
      },
    ]);
    await tick(20_000);
    const done = store.snapshot();
    expect(done).toMatchObject({ phase: "done", early: false });
    expect(done.phase === "done" ? done.calls : []).toEqual([
      {
        entity: "CH-1",
        effector: "chiller__start",
        atS: 600,
        accepted: true,
        reason: "",
        effectEtaNs: 30e9,
        immediate: false,
        atMs: 10_000,
      },
      {
        entity: "CH-1",
        effector: "chiller__stop",
        atS: 900,
        accepted: false,
        reason: "interlock_refused",
        effectEtaNs: null,
        immediate: false,
        atMs: 15_000,
      },
    ]);
  });

  it("never has more than one status request outstanding", async () => {
    const fake = launchBridge("w-4");
    const { store, tick } = await started(fake);
    const gate = fake.holdAction("job.status");
    await tick(500);
    await tick(5000);
    expect(fake.calls.filter((call) => call.name === "job.status")).toHaveLength(1);
    gate.release();
    await flushPromises();
    await tick(500);
    expect(fake.calls.filter((call) => call.name === "job.status")).toHaveLength(2);
    expect(store.snapshot()).toMatchObject({ phase: "running" });
  });
});

describe("launch store: proposals", () => {
  it("clears the proposal when a run starts", async () => {
    const fake = launchBridge("w-4");
    const { store } = launch(fake);
    store.propose("factory", plan);
    expect(store.proposal("factory")).not.toBeNull();
    await store.start(origin, plan);
    expect(store.proposal("factory")).toBeNull();
  });
});
