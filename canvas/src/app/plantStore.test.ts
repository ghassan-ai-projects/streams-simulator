import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { FakeBridge, failure, sampleWorkspace } from "../test/fakeBridge";
import { chillerPorts, composition, domain } from "../test/plantFixture";
import { PlantStore } from "./plantStore";

function file(path: string, value: unknown) {
  return success({ path, content: typeof value === "string" ? value : JSON.stringify(value), etag: "e" });
}

function bridge() {
  const fake = new FakeBridge();
  fake.files.set("domains/mini.domain.json", file("domains/mini.domain.json", domain));
  fake.files.set("compositions/mini.composition.json", file("compositions/mini.composition.json", composition));
  return fake;
}

const workspace = {
  ...sampleWorkspace,
  compositions: [{ path: "compositions/mini.composition.json", id: "mini", title: "Mini" }],
};

describe("plant store", () => {
  it("loads the domain, its composition and each component's ports through the bridge", async () => {
    const fake = bridge();
    fake.actions.set(
      "component.show",
      success({ entry: { class: "static/heat_exchanger/chiller", tags: ["hvac"] }, document: chillerPorts }),
    );
    const store = new PlantStore(fake);
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
    const store = new PlantStore(fake);
    await store.load("domains/mini.domain.json", sampleWorkspace);
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.composition).toBeNull();
    expect(state.specs.get("lib/chiller")).toEqual({ ports: [], category: "env" });
  });

  it("selects and clears an instance only when a plant is open", async () => {
    const store = new PlantStore(bridge());
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
    const store = new PlantStore(fake);
    await store.load("domains/none.domain.json", sampleWorkspace);
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { code: "not_found" } });
    fake.files.set("domains/bad.domain.json", file("domains/bad.domain.json", "{not json"));
    fake.files.set("compositions/mini.composition.json", failed(failure("too_large", 413)));
    await store.load("domains/bad.domain.json", workspace);
    expect(store.snapshot()).toMatchObject({ phase: "ready", plant: { instances: [] } });
  });

  it("keeps the open plant and its selection when the workspace listing changes", async () => {
    const store = new PlantStore(bridge());
    await store.load("domains/mini.domain.json", sampleWorkspace);
    store.select("CH-1");
    const open = store.snapshot();
    await store.load("domains/mini.domain.json", { ...sampleWorkspace, runs: [] });
    expect(store.snapshot()).toBe(open);
  });
});
