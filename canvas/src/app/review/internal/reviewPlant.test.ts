import { describe, expect, it, vi } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { domain, domainDocument } from "../../../test/plantFixture";
import { reviewState } from "../../../test/reviewFixture";
import { report } from "../../../test/runFixture";
import { createPlantSession } from "../../plant";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import type { PlantSession } from "../../plant";
import { reviewPlantFor, showRunPlant } from "./reviewPlant";

const workspace = {
  ...sampleWorkspace,
  domains: [
    { path: "domains/mini.domain.json", id: "mini", title: "Mini on disk", instances: 3 },
    { path: "domains/legacy.domain.json", id: "", title: "", instances: 0 },
  ],
};
const recorded = domainDocument({ id: "mini" });
const idle = { phase: "idle" } as const;

async function shownAs(key: string) {
  const plants = createPlantSession(new FakeBridge().plant);
  await plants.show(key, domainDocument(domain));
  return plants.snapshot();
}

describe("the plant a review draws", () => {
  it("shows the plant the run recorded, titled by it, with the workspace's file to open", async () => {
    const plant = reviewPlantFor(reviewState({}), await shownAs("run:w-1"), workspace);
    expect(plant).toMatchObject({ title: "Mini plant", path: "domains/mini.domain.json", key: "run:w-1" });
    expect(plant.plant).not.toBeNull();
  });

  it("draws the workspace file as the fallback when the recorded plant cannot be read", async () => {
    const state = reviewState({});
    expect(reviewPlantFor(state, await shownAs("domains/mini.domain.json"), workspace).plant).not.toBeNull();
    expect(reviewPlantFor(state, await shownAs("run:w-9"), workspace).plant).toBeNull();
    expect(reviewPlantFor(state, await shownAs("domains/other.domain.json"), workspace)).toMatchObject({
      plant: null,
      fallback: "domains/mini.domain.json",
    });
  });

  it("titles a plant that is not shown by the workspace entry, then the domain id, then Plant", () => {
    const state = reviewState({});
    expect(reviewPlantFor(state, idle, workspace).title).toBe("Mini on disk");
    expect(reviewPlantFor(state, idle, null)).toMatchObject({ title: "mini", path: null, fallback: undefined });
    const unnamed = reviewState({ report: { phase: "ready", value: { ...report, domainId: "" } } });
    expect(reviewPlantFor(unnamed, idle, null).title).toBe("Plant");
  });

  it("looks up no fallback before the report is ready", () => {
    const loading = reviewState({ report: { phase: "loading" } });
    expect(reviewPlantFor(loading, idle, workspace)).toMatchObject({ fallback: undefined, key: "run:w-1" });
  });
});

function session() {
  const show = vi.fn<PlantSession["show"]>();
  const load = vi.fn<PlantSession["load"]>();
  const plants: PlantSession = {
    snapshot: () => idle,
    subscribe: () => () => undefined,
    show,
    load,
    select: vi.fn(),
    close: vi.fn(),
  };
  return { plants, show, load };
}

describe("opening a review's plant", () => {
  const target = { key: "run:w-1", fallback: "domains/mini.domain.json" };

  it("shows the plant the run recorded under the run's key", () => {
    const { plants, show, load } = session();
    showRunPlant(plants, target, { phase: "ready", value: recorded }, workspace);
    expect(show).toHaveBeenCalledWith("run:w-1", recorded);
    expect(load).not.toHaveBeenCalled();
  });

  it("loads the workspace's file when the recorded plant failed and there is one to load", () => {
    const failed = {
      phase: "failed",
      failure: { kind: "missing", diagnostic: { displayCode: "x", message: "m" } },
    } as const;
    const { plants, show, load } = session();
    showRunPlant(plants, target, failed, workspace);
    expect(load).toHaveBeenCalledWith("domains/mini.domain.json", workspace);
    showRunPlant(plants, { key: "run:w-1", fallback: undefined }, failed, workspace);
    showRunPlant(plants, target, failed, null);
    showRunPlant(plants, target, { phase: "loading" }, workspace);
    expect(load).toHaveBeenCalledTimes(1);
    expect(show).not.toHaveBeenCalled();
  });
});
