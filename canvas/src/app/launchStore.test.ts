import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { FakeBridge, failure } from "../test/fakeBridge";
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

const plan = { seed: 3, durationS: 40, explore: false, faults: [{ entity: "CT-1", fault: "fan", atS: 10 }] };

describe("launch", () => {
  it("creates the world, injects at the fault's time, advances to the end and ends the run", async () => {
    const fake = bridge();
    const store = new LaunchStore(fake);
    await store.start("factory", plan);
    expect(store.snapshot()).toEqual({ phase: "done", worldId: "w-4" });
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

  it("reports the binary's refusal and discards the world", async () => {
    const fake = bridge();
    fake.actions.set("fault.inject", failed(failure("unknown_fault", 422, "no such fault")));
    const store = new LaunchStore(fake);
    await store.start("factory", plan);
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { code: "unknown_fault" } });
    expect(fake.calls.at(-1)).toEqual({ name: "world.destroy", args: { world_id: "w-4" } });
  });

  it("stops at the next step when canceled and can be reset", async () => {
    const fake = bridge();
    const store = new LaunchStore(fake);
    const running = store.start("factory", { ...plan, faults: [] });
    await store.start("factory", plan);
    store.cancel();
    expect(store.snapshot()).toMatchObject({ phase: "running", canceling: true });
    await running;
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { code: "canceled" } });
    expect(fake.calls.filter((call) => call.name === "world.create")).toHaveLength(1);
    store.reset();
    store.cancel();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });

  it("fails without a world when the world cannot be created", async () => {
    const fake = bridge();
    fake.actions.set("world.create", failed(failure("domain_invalid", 422, "bad domain")));
    const store = new LaunchStore(fake);
    await store.start("factory", plan);
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { code: "domain_invalid" } });
    expect(fake.calls.map((call) => call.name)).toEqual(["world.create"]);
  });
});
