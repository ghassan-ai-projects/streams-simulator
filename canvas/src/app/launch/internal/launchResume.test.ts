import { describe, expect, it } from "vitest";
import { failure as failed } from "../../ports";
import { cannedFailure } from "../../../test/fakeBridge";
import { flushPromises } from "../../../test/deferred";
import { launchBridge } from "../../../test/launchFixture";
import { launch, origin, started } from "../../../test/launchRun";

describe("launch store: finding a job after a reload", () => {
  const request = {
    domain: "factory",
    seed: 3,
    explore: false,
    duration_s: 1800,
    rate: 60,
    faults: [],
    calls: [{ entity_id: "CH-1", effector: "chiller__start", at_s: 600, args: { chiller_id: "CH-1" } }],
  };
  const originOf = (domain: string) => (domain === "factory" ? origin : null);

  function reloaded(reachedS = 600) {
    const fake = launchBridge("w-4");
    const run = launch(fake);
    fake.jobs.adopt(reachedS, request);
    return { fake, ...run };
  }

  it("takes over a running job: its progress, its outcomes so far, Stop now and the poll", async () => {
    const { store, tick, fake } = reloaded(900);
    await store.resume(originOf);
    const adopted = store.snapshot();
    expect(adopted).toMatchObject({
      phase: "running",
      origin,
      worldId: "w-4",
      atS: 900,
      durationS: 1800,
      rate: 60,
      stopping: false,
      plan: { seed: 3, durationS: 1800, explore: false, faults: [], perturbations: [], calls: [], recovered: true },
    });
    expect(adopted.phase === "running" ? adopted.calls : []).toMatchObject([
      { effector: "chiller__start", accepted: true, atMs: Number.NEGATIVE_INFINITY },
    ]);
    await tick(14_500);
    expect(store.snapshot()).toMatchObject({ phase: "running", atS: 1770 });
    await tick(500);
    expect(store.snapshot()).toMatchObject({ phase: "done", early: false, reachedS: 1800, plan: { recovered: true } });
    expect(fake.jobs.requests).toEqual([]);
  });

  describe("with the plan the director returns", () => {
    const planned = {
      ...request,
      name: "Fan failure",
      purpose: "Observe recovery",
      faults: [{ entity_id: "CT-1", fault: "fan", at_s: 600, until_s: 900 }],
      perturbations: [{ kind: "drop", params: { rate: 0.5 }, at_s: 300 }],
    };

    it("rebuilds the faults, the delivery conditions, the name and the purpose of an exploration", async () => {
      const fake = launchBridge("w-4");
      const run = launch(fake);
      fake.jobs.adopt(700, { ...planned, explore: true });
      await run.store.resume(originOf);
      expect(run.store.snapshot()).toMatchObject({
        phase: "running",
        plan: {
          name: "Fan failure",
          purpose: "Observe recovery",
          explore: true,
          faults: [{ entity: "CT-1", fault: "fan", atS: 600, untilS: 900 }],
          perturbations: [{ kind: "drop", params: { rate: 0.5 }, atS: 300 }],
          recovered: true,
        },
      });
    });

    it("keeps an evaluation's plan sealed, naming the run all the same", async () => {
      const fake = launchBridge("w-4");
      const run = launch(fake);
      fake.jobs.adopt(700, planned);
      await run.store.resume(originOf);
      const state = run.store.snapshot();
      expect(state).toMatchObject({ phase: "running", plan: { name: "Fan failure", faults: [], perturbations: [] } });
      expect(JSON.stringify(state)).not.toContain("fan");
      expect(JSON.stringify(state)).not.toContain("drop");
    });
  });

  it("can stop an adopted job", async () => {
    const { store, tick, fake } = reloaded();
    await store.resume(originOf);
    store.stopEarly();
    await flushPromises();
    await tick(500);
    expect(fake.calls.some((call) => call.name === "job.stop")).toBe(true);
    expect(store.snapshot()).toMatchObject({ phase: "done", early: true, reachedS: 600 });
  });

  it("names the plant by its id when the workspace no longer lists it", async () => {
    const { store } = reloaded();
    await store.resume(() => null);
    expect(store.snapshot()).toMatchObject({
      phase: "running",
      origin: { domain: "factory", title: "factory", path: "" },
    });
  });

  it("does not bring back a job that already ended while the page was away", async () => {
    const { store, tick } = reloaded();
    await tick(40_000);
    await store.resume(originOf);
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });

  it("leaves things alone when there is no job, or the listing fails, or a run is already going", async () => {
    const none = launch(launchBridge("w-4"));
    await none.store.resume(originOf);
    expect(none.store.snapshot()).toEqual({ phase: "idle" });
    const fake = launchBridge("w-4");
    fake.actions.set("job.list", failed(cannedFailure("bridge_failed", 500, "down")));
    const broken = launch(fake);
    await broken.store.resume(originOf);
    expect(broken.store.snapshot()).toEqual({ phase: "idle" });
    const busy = await started();
    const before = busy.fake.calls.length;
    await busy.store.resume(originOf);
    expect(busy.fake.calls.length).toBe(before);
  });

  it("shows an adopted job that is already stopping as stopping", async () => {
    const { store, fake } = reloaded();
    fake.jobs.answer("job.stop", { world_id: "w-4" });
    await store.resume(originOf);
    expect(store.snapshot()).toMatchObject({ phase: "running", stopping: true });
  });
});
