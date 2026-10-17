import { createLaunchStore } from "../app/launch";
import { success } from "../app/ports";
import { FakeBridge } from "./fakeBridge";
import type { Canned } from "./fakeFailures";
import { fakeClock } from "./fakeClock";

export const advanced = success({ clock: "2026-01-01T00:30:00Z", emitted: 12 });

export function launchActions(worldId = "w-7"): ReadonlyMap<string, Canned<unknown>> {
  return new Map<string, Canned<unknown>>([
    ["world.create", success({ world_id: worldId })],
    ["clock.advance", advanced],
    ["fault.inject", success({ fault_id: "f-0", simulated: true })],
    ["run.end", success({ run_artifact_path: `runs/${worldId}/run.json` })],
    ["world.destroy", success({ destroyed: true })],
  ]);
}

export function launchBridge(worldId = "w-7") {
  const fake = new FakeBridge();
  launchActions(worldId).forEach((outcome, name) => fake.actions.set(name, outcome));
  return fake;
}

export function launchHarness(bridge: FakeBridge) {
  const timers = fakeClock();
  return { launch: createLaunchStore(bridge.launch, timers, timers), timers };
}
