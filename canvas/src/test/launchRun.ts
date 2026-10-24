import type { LaunchPlan } from "../model/launch";
import { createLaunchStore } from "../app/launch";
import type { FakeBridge } from "./fakeBridge";
import { fakeClock } from "./fakeClock";
import { elapse, launchBridge } from "./launchFixture";

export const origin = { domain: "factory", title: "Factory", path: "domains/factory.domain.json" };

export const plan: LaunchPlan = {
  seed: 3,
  durationS: 1800,
  explore: false,
  pace: "watch",
  faults: [{ entity: "CT-1", fault: "fan", atS: 600 }],
  calls: [],
};

export const start = { entity: "CH-1", effector: "chiller__start", atS: 600, args: { chiller_id: "CH-1" } };

export function launch(fake: FakeBridge) {
  const timers = fakeClock();
  fake.jobs.drive(timers);
  const store = createLaunchStore(fake.launch, timers, timers);
  const tick = (ms: number) => elapse(timers, ms);
  return { store, timers, tick };
}

export async function started(fake = launchBridge("w-4"), launched = plan) {
  const run = launch(fake);
  await run.store.start(origin, launched);
  return { ...run, fake };
}
