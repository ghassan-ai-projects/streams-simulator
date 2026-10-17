import { describe, expect, it, vi } from "vitest";
import { failure as failed, success } from "../../ports";
import { FakeBridge, failure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { LaunchStore } from "./launchStore";

function bridge() {
  const fake = new FakeBridge();
  fake.actions.set("world.create", success({ world_id: "w-4" }));
  fake.actions.set("clock.advance", success({}));
  fake.actions.set("fault.inject", success({ fault_id: "f-0" }));
  fake.actions.set("run.end", success({ run_artifact_path: "runs/w-4/run.json" }));
  fake.actions.set("world.destroy", success({}));
  return fake;
}

const origin = { domain: "factory", title: "Factory", path: "domains/factory.domain.json" };
const plan = {
  seed: 3,
  durationS: 40,
  explore: false,
  faults: [{ entity: "CT-1", fault: "fan", atS: 10 }],
  calls: [],
};

function launch(fake: FakeBridge): LaunchStore {
  const timers = fakeClock();
  return new LaunchStore(fake.launch, timers, timers);
}

describe("launch", () => {
  it("creates the world, injects at the fault's time, advances to the end and ends the run", async () => {
    const fake = bridge();
    const store = launch(fake);
    await store.start(origin, plan);
    expect(store.snapshot()).toEqual({
      phase: "done",
      origin,
      worldId: "w-4",
      early: false,
      plan,
      reachedS: plan.durationS,
      calls: [],
    });
    const names = fake.calls.map((call) => call.name);
    expect([names[0], names.at(-1)]).toEqual(["world.create", "run.end"]);
    const advanced = (calls: typeof fake.calls) =>
      calls.filter((call) => call.name === "clock.advance").reduce((sum, call) => sum + Number(call.args.by_ns), 0);
    const inject = names.indexOf("fault.inject");
    expect(advanced(fake.calls.slice(0, inject))).toBe(10e9);
    expect(advanced(fake.calls)).toBe(40e9);
    expect(fake.calls[0]?.args).toEqual({ domain: "factory", seed: 3, explore: false });
    expect(fake.calls[inject]?.args).toEqual({ world_id: "w-4", entity_id: "CT-1", fault: "fan" });
  });

  it("invokes an effector at its minute, keeps a refusal as an outcome and goes on", async () => {
    const fake = bridge();
    fake.handlers.set("effector.invoke", (args) =>
      args.effector === "chiller__stop"
        ? success({ accepted: false, reason: "interlock_refused", world_id: "w-4", command_id: "director-2" })
        : success({ accepted: true, effect_eta_ns: 30e9, world_id: "w-4", command_id: "director-1" }),
    );
    const start = { entity: "CH-1", effector: "chiller__start", atS: 10, args: { chiller_id: "CH-1" } };
    const stop = { ...start, effector: "chiller__stop", atS: 20 };
    const store = launch(fake);
    await store.start(origin, { ...plan, faults: [], calls: [stop, start] });
    const state = store.snapshot();
    expect(state).toMatchObject({ phase: "done", early: false });
    expect(state.phase === "done" ? state.calls : []).toEqual([
      {
        entity: "CH-1",
        effector: "chiller__start",
        atS: 10,
        accepted: true,
        reason: "",
        effectEtaNs: 30e9,
        atMs: expect.any(Number) as number,
      },
      {
        entity: "CH-1",
        effector: "chiller__stop",
        atS: 20,
        accepted: false,
        reason: "interlock_refused",
        effectEtaNs: null,
        atMs: expect.any(Number) as number,
      },
    ]);
    const invoked = fake.calls.filter((call) => call.name === "effector.invoke");
    expect(invoked.map((call) => call.args)).toEqual([
      { world_id: "w-4", entity_id: "CH-1", effector: "chiller__start", args: { chiller_id: "CH-1" } },
      { world_id: "w-4", entity_id: "CH-1", effector: "chiller__stop", args: { chiller_id: "CH-1" } },
    ]);
    const at = fake.calls.findIndex((call) => call.name === "effector.invoke");
    const before = fake.calls.slice(0, at).filter((call) => call.name === "clock.advance");
    expect(before.reduce((sum, call) => sum + Number(call.args.by_ns), 0)).toBe(10e9);
  });

  it("publishes the call the moment it is made, stamped with the wall clock, so the plant can hold its mark", async () => {
    const fake = bridge();
    fake.actions.set("effector.invoke", success({ accepted: true, effect_eta_ns: 1e9 }));
    let now = 1000;
    const timers = fakeClock();
    const store = new LaunchStore(fake.launch, { nowMs: () => (now += 250) }, timers);
    const seen: { atS: number; calls: number; atMs: number }[] = [];
    store.subscribe(() => {
      const state = store.snapshot();
      if (state.phase === "running") seen.push({ atS: state.atS, calls: state.calls.length, atMs: state.atMs });
    });
    const call = { entity: "CH-1", effector: "chiller__start", atS: 20, args: {} };
    await store.start(origin, { ...plan, faults: [], calls: [call] });
    const first = seen.findIndex((state) => state.calls === 1);
    expect(seen[first]?.atS).toBe(20);
    expect(seen[first - 1]?.calls).toBe(0);
    expect(seen[first]?.atMs).toBeGreaterThan(seen[first - 1]?.atMs ?? 0);
  });

  it("abandons the run when the bridge fails an effector call", async () => {
    const fake = bridge();
    fake.actions.set("effector.invoke", failed(failure("invalid_args", 422, "setpoint out of range")));
    const store = launch(fake);
    const call = { entity: "CH-1", effector: "chiller__start", atS: 0, args: {} };
    await store.start(origin, { ...plan, faults: [], calls: [call] });
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      failure: { kind: "rejected", step: "invoke", diagnostic: { displayCode: "invalid_args" } },
      kept: "w-4",
    });
    expect(fake.calls.at(-1)).toEqual({ name: "world.destroy", args: { world_id: "w-4" } });
  });

  it("reports the binary's refusal and discards the world", async () => {
    const fake = bridge();
    fake.actions.set("fault.inject", failed(failure("unknown_fault", 422, "no such fault")));
    const store = launch(fake);
    await store.start(origin, plan);
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      failure: { kind: "rejected", step: "inject", diagnostic: { displayCode: "unknown_fault" } },
    });
    expect(fake.calls.at(-1)).toEqual({ name: "world.destroy", args: { world_id: "w-4" } });
  });

  it("stops early after at least one step, keeps the shorter run and can be reset", async () => {
    const fake = bridge();
    const store = launch(fake);
    const running = store.start(origin, { ...plan, faults: [], calls: [] });
    await store.start(origin, plan);
    store.stopEarly();
    expect(store.snapshot()).toMatchObject({ phase: "running", stopping: true });
    await running;
    expect(store.snapshot()).toMatchObject({ phase: "done", worldId: "w-4", early: true, calls: [] });
    expect(fake.calls.map((call) => call.name)).toEqual(["world.create", "clock.advance", "run.end"]);
    store.reset();
    store.stopEarly();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });

  it("fails without a world when the world cannot be created", async () => {
    const fake = bridge();
    fake.actions.set("world.create", failed(failure("domain_invalid", 422, "bad domain")));
    const store = launch(fake);
    await store.start(origin, plan);
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      failure: { kind: "rejected", step: "create", diagnostic: { displayCode: "domain_invalid" } },
    });
    expect(fake.calls.map((call) => call.name)).toEqual(["world.create"]);
  });

  it("paces a watched run to the simulated time reached and reports the speed of the last seconds", async () => {
    const fake = bridge();
    const timers = fakeClock();
    const waits: number[] = [];
    const scheduler = {
      schedule: (delayMs: number, task: () => void) => timers.schedule(delayMs, task),
      delay: (ms: number) => {
        waits.push(ms);
        timers.advance(ms);
        return Promise.resolve();
      },
    };
    const store = new LaunchStore(fake.launch, timers, scheduler);
    const rates: (number | null)[] = [];
    store.subscribe(() => {
      const state = store.snapshot();
      if (state.phase === "running") rates.push(state.rate);
    });
    await store.start(origin, { ...plan, durationS: 1800, faults: [], pace: "watch" });
    expect(waits.reduce((sum, ms) => sum + ms, 0)).toBe(30_000);
    expect(waits).toHaveLength(30);
    expect(rates.at(-1)).toBe(60);
    expect(store.snapshot()).toMatchObject({ phase: "done", early: false });
  });

  it("runs a plan without a pace as fast as the binary answers", async () => {
    const waits: number[] = [];
    const timers = fakeClock();
    const scheduler = {
      schedule: (delayMs: number, task: () => void) => timers.schedule(delayMs, task),
      delay: (ms: number) => {
        waits.push(ms);
        return Promise.resolve();
      },
    };
    const store = new LaunchStore(bridge().launch, timers, scheduler);
    await store.start(origin, plan);
    expect(waits).toEqual([]);
  });

  it("wakes a waiting run the moment Stop now is pressed", async () => {
    const timers = fakeClock();
    const store = new LaunchStore(bridge().launch, timers, timers);
    const running = store.start(origin, { ...plan, durationS: 1800, faults: [], pace: "watch" });
    await vi.waitFor(() => {
      expect(store.snapshot()).toMatchObject({ phase: "running", atS: 60 });
    });
    store.stopEarly();
    await running;
    expect(store.snapshot()).toMatchObject({ phase: "done", early: true, reachedS: 60 });
  });
});
