import { createLaunchStore } from "../app/launch";
import { FakeBridge } from "./fakeBridge";
import { flushPromises } from "./deferred";
import type { FakeClock } from "./fakeClock";
import { fakeClock } from "./fakeClock";

export function launchBridge(worldId = "w-7") {
  const fake = new FakeBridge();
  fake.jobs.worldId = worldId;
  return fake;
}

export function launchHarness(bridge: FakeBridge) {
  const timers = fakeClock();
  bridge.jobs.drive(timers);
  return { launch: createLaunchStore(bridge.launch, timers, timers), timers };
}

export async function elapse(timers: FakeClock, ms: number): Promise<void> {
  for (let elapsed = 0; elapsed < ms; elapsed += 500) {
    timers.advance(500);
    await flushPromises();
  }
}
