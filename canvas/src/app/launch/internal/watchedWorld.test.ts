import { idleControl } from "../../../model/launch";
import { describe, expect, it } from "vitest";
import { offeredSession } from "../../../test/interruptFixture";
import { plan, origin as launched } from "../../../test/launchRun";
import type { LaunchState } from "./launchState";
import { followingOf, watchedWorld } from "./watchedWorld";

const running = (worldId: string | null, domain = "factory"): LaunchState => ({
  phase: "running",
  origin: { ...launched, domain },
  worldId,
  atS: 10,
  durationS: 600,
  plan,
  startedMs: 0,
  atMs: 0,
  stopping: false,
  paused: false,
  control: idleControl,
  injected: [],
  rate: null,
  calls: [],
});

describe("the world a plant should watch", () => {
  it("is the world of the job running on that plant", () => {
    expect(watchedWorld(running("w-7"), "factory")).toBe("w-7");
  });

  it("is none before the job has a world, or for another plant", () => {
    expect(watchedWorld(running(null), "factory")).toBeNull();
    expect(watchedWorld(running("w-7", "mini"), "factory")).toBeNull();
    expect(watchedWorld(running("w-7"), "")).toBeNull();
  });

  it("is none once the job is over or when nothing runs", () => {
    expect(watchedWorld({ phase: "idle" }, "factory")).toBeNull();
    const done: LaunchState = {
      phase: "done",
      origin: launched,
      worldId: "w-7",
      early: false,
      plan,
      reachedS: 600,
      calls: [],
    };
    expect(watchedWorld(done, "factory")).toBeNull();
  });
});

describe("what a plant follows", () => {
  const session = { ...offeredSession, domain: "factory", run: "c-1" };

  it("is the connected session on that plant before any job", () => {
    expect(followingOf(session, running("w-7"), "factory")).toEqual({ kind: "session", run: "c-1" });
  });

  it("is the job running on that plant when the session is elsewhere or absent", () => {
    expect(followingOf(session, running("w-7", "mini"), "mini")).toEqual({ kind: "watch", run: "w-7" });
    expect(followingOf({ phase: "idle" }, running("w-7"), "factory")).toEqual({ kind: "watch", run: "w-7" });
  });

  it("is nothing while no plant is ready", () => {
    expect(followingOf(session, running("w-7"), undefined)).toEqual({ kind: "session", run: null });
  });
});
