import { describe, expect, it, vi } from "vitest";
import { fakeClock } from "../../../test/fakeClock";
import { elapse, launchBridge } from "../../../test/launchFixture";
import { LaunchStore } from "./launchStore";

const origin = { domain: "factory", title: "Factory", path: "domains/factory.domain.json" };
const plan = { seed: 3, durationS: 120, explore: false, faults: [], calls: [] };

function setup() {
  const timers = fakeClock();
  const fake = launchBridge("w-4");
  fake.jobs.drive(timers);
  return { timers, launch: new LaunchStore(fake.launch, timers, timers) };
}

describe("run watch", () => {
  it("holds a run it watched for two seconds, then resets the store and opens the run", async () => {
    const { launch, timers } = setup();
    const watch = launch.watch("factory");
    const then = vi.fn<(worldId: string) => void>();
    const release = watch.holdFinished(then);
    await launch.start(origin, plan);
    await elapse(timers, 500);
    expect(watch.shows(launch.snapshot())).toBe(true);
    timers.advance(1999);
    expect(then).not.toHaveBeenCalled();
    timers.advance(1);
    expect(then).toHaveBeenCalledWith("w-4");
    expect(launch.snapshot().phase).toBe("idle");
    release();
  });

  it("does not hold a run it did not watch, and leaves it to be shown as a form", async () => {
    const { launch, timers } = setup();
    await launch.start(origin, plan);
    await elapse(timers, 500);
    const watch = launch.watch("factory");
    const then = vi.fn<(worldId: string) => void>();
    const release = watch.holdFinished(then);
    expect(watch.shows(launch.snapshot())).toBe(false);
    timers.advance(10_000);
    expect(then).not.toHaveBeenCalled();
    expect(launch.snapshot().phase).toBe("done");
    release();
  });

  it("watches a run that is already going when it is made", async () => {
    const { launch, timers } = setup();
    const running = launch.start(origin, plan);
    const watch = launch.watch("factory");
    const then = vi.fn<(worldId: string) => void>();
    const release = watch.holdFinished(then);
    await running;
    await elapse(timers, 500);
    timers.advance(2000);
    expect(then).toHaveBeenCalledWith("w-4");
    release();
  });

  it("ignores the runs of another plant", async () => {
    const { launch, timers } = setup();
    const watch = launch.watch("elsewhere");
    const then = vi.fn<(worldId: string) => void>();
    const release = watch.holdFinished(then);
    await launch.start(origin, plan);
    await elapse(timers, 500);
    expect(watch.shows(launch.snapshot())).toBe(false);
    timers.advance(10_000);
    expect(then).not.toHaveBeenCalled();
    release();
  });

  it("shows the result of a run with effector calls and leaves it for the person to open", async () => {
    const timers = fakeClock();
    const fake = launchBridge("w-4");
    fake.jobs.drive(timers);
    const launch = new LaunchStore(fake.launch, timers, timers);
    const watch = launch.watch("factory");
    const then = vi.fn<(worldId: string) => void>();
    const release = watch.holdFinished(then);
    await launch.start(origin, { ...plan, calls: [{ entity: "CH-1", effector: "start", atS: 0, args: {} }] });
    await elapse(timers, 500);
    expect(watch.shows(launch.snapshot())).toBe(true);
    timers.advance(10_000);
    expect(then).not.toHaveBeenCalled();
    release();
  });

  it("stops holding when released", async () => {
    const { launch, timers } = setup();
    const watch = launch.watch("factory");
    const then = vi.fn<(worldId: string) => void>();
    const release = watch.holdFinished(then);
    await launch.start(origin, plan);
    await elapse(timers, 500);
    release();
    timers.advance(10_000);
    expect(then).not.toHaveBeenCalled();
    expect(timers.pending()).toBe(0);
  });
});
