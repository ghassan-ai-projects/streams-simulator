import { describe, expect, it } from "vitest";
import { flushPromises } from "../../../test/deferred";
import { failure as failed, success } from "../../ports";
import { FakeBridge, cannedFailure, cannedFile } from "../../../test/fakeBridge";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { chillerPorts, componentShown, composition, domain, domainDocument } from "../../../test/plantFixture";
import { createPlantSession } from "./plantSession";

function bridge() {
  const fake = new FakeBridge();
  fake.addFile("domains/mini.domain.json", domain);
  fake.addFile("compositions/mini.composition.json", composition);
  return fake;
}

const workspace = {
  ...sampleWorkspace,
  compositions: [{ path: "compositions/mini.composition.json", id: "mini", title: "Mini", instances: 0 }],
};

describe("plant store", () => {
  it("loads the domain, its composition and each component's ports through the bridge", async () => {
    const fake = bridge();
    fake.actions.set(
      "component.show",
      success(componentShown(chillerPorts, {}, { class: "static/heat_exchanger/chiller", tags: ["hvac"] })),
    );
    const store = createPlantSession(fake.plant);
    await store.load("domains/mini.domain.json", workspace);
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.composition).toBe("compositions/mini.composition.json");
    expect(state.plant.links[0]?.fromPort).toBe("heat_out");
    expect(fake.calls.map((call) => call.args)).toEqual([{ ref: "lib/chiller" }, { ref: "lib/tower" }]);
    expect(state.specs.get("lib/tower")?.ports.map((port) => port.name)).toEqual([
      "cond_in",
      "spare_in",
      "heat_out",
      "power_out",
    ]);
  });

  it("opens a domain without a composition, and tolerates a library that refuses", async () => {
    const fake = bridge();
    const store = createPlantSession(fake.plant);
    await store.load("domains/mini.domain.json", sampleWorkspace);
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.composition).toBeNull();
    expect(state.specs.get("lib/chiller")).toEqual({ ports: [], category: "env" });
  });

  it("selects and clears an instance only when a plant is open", async () => {
    const store = createPlantSession(bridge().plant);
    store.select("CH-1");
    expect(store.snapshot()).toEqual({ phase: "idle" });
    await store.load("domains/mini.domain.json", sampleWorkspace);
    store.select("CH-1");
    expect(store.snapshot()).toMatchObject({ selected: "CH-1" });
    store.select(null);
    expect(store.snapshot()).toMatchObject({ selected: null });
  });

  it("reports a missing file and survives a file that is not JSON", async () => {
    const fake = bridge();
    const store = createPlantSession(fake.plant);
    await store.load("domains/none.domain.json", sampleWorkspace);
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      failure: { kind: "missing", diagnostic: { displayCode: "not_found" } },
    });
    fake.addFile("domains/bad.domain.json", "{not json");
    fake.files.set("compositions/mini.composition.json", failed(cannedFailure("too_large", 413)));
    await store.load("domains/bad.domain.json", workspace);
    expect(store.snapshot()).toMatchObject({ phase: "ready", plant: { instances: [] } });
  });

  it("keeps the open plant and its selection when the workspace listing changes", async () => {
    const store = createPlantSession(bridge().plant);
    await store.load("domains/mini.domain.json", sampleWorkspace);
    store.select("CH-1");
    const open = store.snapshot();
    await store.load("domains/mini.domain.json", { ...sampleWorkspace, runs: [] });
    expect(store.snapshot()).toBe(open);
  });

  it("reloads the open plant when its file changed on disk and keeps a selection that still exists", async () => {
    const fake = bridge();
    const store = createPlantSession(fake.plant);
    await store.load("domains/mini.domain.json", sampleWorkspace);
    store.select("CT-1");
    const grown = { ...domain, instances: [...domain.instances, { kind: "meter", names: ["MB-2"], area: "Power" }] };
    fake.addFile("domains/mini.domain.json", grown, "e2");
    await store.load("domains/mini.domain.json", sampleWorkspace);
    expect(store.snapshot()).toMatchObject({ phase: "ready", selected: "CT-1" });
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.plant.instances.map((instance) => instance.id)).toContain("MB-2");
  });

  it("reloads the open plant when its composition appears in the listing", async () => {
    const fake = bridge();
    const store = createPlantSession(fake.plant);
    await store.load("domains/mini.domain.json", sampleWorkspace);
    expect(store.snapshot()).toMatchObject({ composition: null });
    await store.load("domains/mini.domain.json", workspace);
    expect(store.snapshot()).toMatchObject({ phase: "ready", composition: "compositions/mini.composition.json" });
  });

  it("ends on the plant asked for last when two loads overlap", async () => {
    const fake = bridge();
    fake.addFile("domains/other.domain.json", { ...domain, id: "other" });
    const held = fake.holdRead("domains/mini.domain.json");
    const store = createPlantSession(fake.plant);
    const first = store.load("domains/mini.domain.json", workspace);
    const second = store.load("domains/other.domain.json", workspace);
    await second;
    held[0]?.resolve(cannedFile("domains/mini.domain.json", domain));
    await first;
    expect(store.snapshot()).toMatchObject({ phase: "ready", path: "domains/other.domain.json" });
  });

  it("reads once while a load of the same plant is in flight, and shows a plant without reading", async () => {
    const fake = bridge();
    const held = fake.holdRead("domains/mini.domain.json", 1);
    const store = createPlantSession(fake.plant);
    const first = store.load("domains/mini.domain.json", workspace);
    const second = store.load("domains/mini.domain.json", workspace);
    await flushPromises();
    held[0]?.resolve(cannedFile("domains/mini.domain.json", domain));
    await Promise.all([first, second]);
    expect(fake.reads.filter((path) => path === "domains/mini.domain.json")).toHaveLength(1);
    await store.show("memory:mini", domainDocument(domain));
    expect(fake.reads.filter((path) => path === "memory:mini")).toHaveLength(0);
    expect(store.snapshot()).toMatchObject({ phase: "ready", path: "memory:mini", composition: null });
  });

  it("pairs the composition of the workspace asked for last when the same plant is asked for twice", async () => {
    const fake = bridge();
    const held = fake.holdRead("domains/mini.domain.json", 1);
    const store = createPlantSession(fake.plant);
    const first = store.load("domains/mini.domain.json", sampleWorkspace);
    const second = store.load("domains/mini.domain.json", workspace);
    await flushPromises();
    held[0]?.resolve(cannedFile("domains/mini.domain.json", domain));
    await Promise.all([first, second]);
    expect(store.snapshot()).toMatchObject({ phase: "ready", composition: "compositions/mini.composition.json" });
  });

  it("ignores an earlier visit's answer when the same plant is asked for again", async () => {
    const fake = bridge();
    fake.addFile("domains/other.domain.json", { ...domain, id: "other" });
    const held = fake.holdRead("domains/mini.domain.json");
    const store = createPlantSession(fake.plant);
    const first = store.load("domains/mini.domain.json", workspace);
    await store.load("domains/other.domain.json", workspace);
    const third = store.load("domains/mini.domain.json", workspace);
    held[0]?.resolve(cannedFile("domains/mini.domain.json", { ...domain, title: "stale" }));
    await first;
    expect(store.snapshot()).not.toMatchObject({ phase: "ready", plant: { title: "stale" } });
    held[1]?.resolve(cannedFile("domains/mini.domain.json", domain));
    await third;
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.plant.title).toBe(domain.title);
  });

  it("rejects a late load after close and accepts a fresh close-open visit", async () => {
    const fake = bridge();
    const held = fake.holdRead("domains/mini.domain.json");
    const store = createPlantSession(fake.plant);
    const first = store.load("domains/mini.domain.json", workspace);
    store.close();
    const reopened = store.load("domains/mini.domain.json", workspace);
    held[0]?.resolve(cannedFile("domains/mini.domain.json", { ...domain, title: "late" }));
    await first;
    expect(store.snapshot()).toMatchObject({ phase: "loading", path: "domains/mini.domain.json" });
    held[1]?.resolve(cannedFile("domains/mini.domain.json", domain));
    await reopened;
    expect(store.snapshot()).toMatchObject({ phase: "ready", plant: { title: domain.title } });
  });
});
