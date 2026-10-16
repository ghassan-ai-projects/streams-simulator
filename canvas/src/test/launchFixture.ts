import { success } from "../model/bridge";
import { FakeBridge } from "./fakeBridge";
import { fakeClock } from "./fakeClock";
import { LaunchStore } from "../app/launchStore";

export function launchBridge() {
  const fake = new FakeBridge();
  fake.actions.set("world.create", success({ world_id: "w-7" }));
  fake.actions.set("clock.advance", success({}));
  fake.actions.set("fault.inject", success({ fault_id: "f-0" }));
  fake.actions.set("run.end", success({}));
  fake.actions.set("world.destroy", success({}));
  return fake;
}

export function launchHarness(bridge: FakeBridge) {
  const timers = fakeClock();
  return { launch: new LaunchStore(bridge, timers, timers), scheduler: timers, timers };
}
