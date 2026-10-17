import { LaunchStore } from "../app/launch";
import { success } from "../app/ports";
import { FakeBridge } from "./fakeBridge";
import { fakeClock } from "./fakeClock";

export const advanced = success({ clock: "2026-01-01T00:30:00Z", emitted: 12 });

export function launchBridge(worldId = "w-7") {
  const fake = new FakeBridge();
  fake.actions.set("world.create", success({ world_id: worldId }));
  fake.actions.set("clock.advance", advanced);
  fake.actions.set("fault.inject", success({ fault_id: "f-0", simulated: true }));
  fake.actions.set("run.end", success({ run_artifact_path: `runs/${worldId}/run.json` }));
  fake.actions.set("world.destroy", success({ destroyed: true }));
  return fake;
}

export function launchHarness(bridge: FakeBridge) {
  const timers = fakeClock();
  return { launch: new LaunchStore(bridge.launch, timers, timers), timers };
}
