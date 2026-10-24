import { describe, expect, it } from "vitest";
import { plan, origin as launched } from "../../../test/launchRun";
import type { LaunchState } from "./launchState";
import { watchedWorld } from "./watchedWorld";

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
