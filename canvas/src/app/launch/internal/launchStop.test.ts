import { describe, expect, it } from "vitest";
import { failure as failed } from "../../ports";
import { cannedFailure } from "../../../test/fakeBridge";
import { flushPromises } from "../../../test/deferred";
import { launchBridge } from "../../../test/launchFixture";
import { launch, origin, plan, started } from "../../../test/launchRun";

describe("launch store: stopping", () => {
  it("asks the binary to stop, shows it stopping, and keeps the shorter run once the job says stopped", async () => {
    const { fake, store, tick, timers } = await started();
    await tick(10_000);
    store.stopEarly();
    expect(store.snapshot()).toMatchObject({ phase: "running", stopping: true });
    await flushPromises();
    expect(fake.calls.find((call) => call.name === "job.stop")?.args).toEqual({ world_id: "w-4" });
    expect(store.snapshot()).toMatchObject({ phase: "running", stopping: true, atS: 600 });
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "done", worldId: "w-4", early: true, reachedS: 600, calls: [] });
    expect(timers.pending()).toBe(0);
    store.reset();
    store.stopEarly();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });

  it("asks only once however often Stop now is pressed", async () => {
    const { fake, store, tick } = await started();
    await tick(500);
    store.stopEarly();
    store.stopEarly();
    await flushPromises();
    expect(fake.calls.filter((call) => call.name === "job.stop")).toHaveLength(1);
  });

  it("stops a run the moment the binary has answered, when Stop now came before its answer", async () => {
    const fake = launchBridge("w-4");
    const gate = fake.holdAction("job.start");
    const { store } = launch(fake);
    const starting = store.start(origin, plan);
    store.stopEarly();
    expect(store.snapshot()).toMatchObject({ phase: "running", worldId: null, stopping: true });
    gate.release();
    await starting;
    await flushPromises();
    expect(fake.calls.map((call) => call.name)).toEqual(["job.start", "job.stop"]);
  });

  it("lets the user try again when the binary refuses to stop", async () => {
    const fake = launchBridge("w-4");
    const { store } = await started(fake);
    fake.actions.set("job.stop", failed(cannedFailure("bridge_failed", 500, "busy")));
    store.stopEarly();
    await flushPromises();
    expect(store.snapshot()).toMatchObject({ phase: "running", stopping: false });
  });
});

describe("launch store: when the job fails or contact is lost", () => {
  it("shows the job's own reason and keeps the run that was recorded", async () => {
    const { fake, store, tick } = await started();
    await tick(10_000);
    fake.jobs.fail("job failed: effector.invoke: unknown effector");
    await tick(500);
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      origin,
      kept: "w-4",
      failure: { kind: "rejected", step: "run", diagnostic: { displayCode: "job_failed" } },
      words: "job failed: effector.invoke: unknown effector. Check the plan and start again.",
    });
  });

  it("puts up with a few unanswered polls, then says contact was lost while the run may go on", async () => {
    const { fake, store, tick } = await started();
    fake.actions.set("job.status", failed(cannedFailure("bridge_failed", 503, "down")));
    await tick(500);
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "running" });
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "failed", kept: "w-4", failure: { kind: "unavailable" } });
    expect(store.snapshot()).toMatchObject({ words: expect.stringContaining("lost contact") as string });
  });

  it("forgets the missed polls once one is answered", async () => {
    const { fake, store, tick } = await started();
    fake.actions.set("job.status", failed(cannedFailure("bridge_failed", 503, "down")));
    await tick(500);
    await tick(500);
    fake.actions.delete("job.status");
    await tick(500);
    fake.actions.set("job.status", failed(cannedFailure("bridge_failed", 503, "down")));
    await tick(500);
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "running" });
  });

  it("fails at once when the binary no longer knows the job", async () => {
    const { fake, store, tick } = await started();
    fake.actions.set("job.status", failed(cannedFailure("world_not_found", 404, "no job for world w-4")));
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "failed", kept: "w-4", failure: { kind: "rejected" } });
  });

  it("cannot be reset while running, and can be once it has failed", async () => {
    const { store } = await started();
    store.reset();
    expect(store.snapshot().phase).toBe("running");
    const fake = launchBridge("w-4");
    fake.actions.set("job.start", failed(cannedFailure("job_active", 409, "busy")));
    const next = launch(fake);
    await next.store.start(origin, plan);
    next.store.reset();
    expect(next.store.snapshot()).toEqual({ phase: "idle" });
  });
});
