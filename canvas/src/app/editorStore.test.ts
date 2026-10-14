import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { equipmentOf } from "../model/composition";
import { FakeBridge, failure } from "../test/fakeBridge";
import { EditorStore } from "./editorStore";

const path = "compositions/lab.composition.json";
const doc = {
  id: "lab",
  title: "Lab",
  areas: [{ name: "Site" }],
  equipment: [{ use: "streamsim/chiller@1", names: ["CH-1"], area: "Site" }],
  connections: [],
};

function bridge() {
  const fake = new FakeBridge();
  fake.files.set(path, success({ path, content: JSON.stringify(doc), etag: "c1" }));
  fake.files.set("domains/lab.domain.json", success({ path: "domains/lab.domain.json", content: "{}", etag: "d1" }));
  fake.actions.set("component.list", success({ components: [{ id: "streamsim/pump", kind: "equipment" }] }));
  fake.actions.set("component.show", success({ document: { presets: { small: {} }, default_preset: "small" } }));
  fake.actions.set("domain.compose", success({ channels: 3, faults: 2, effectors: 1 }));
  return fake;
}

async function opened(fake = bridge()) {
  const store = new EditorStore(fake);
  await store.open(path);
  const state = store.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return { store, fake, state };
}

describe("editor store", () => {
  it("opens a composition with the library and its components' ports", async () => {
    const { state, fake } = await opened();
    expect(state.library.map((entry) => entry.ref)).toEqual(["streamsim/pump@1"]);
    expect([...state.specs.keys()]).toEqual(["streamsim/chiller"]);
    expect(fake.calls.at(-1)).toEqual({ name: "component.show", args: { ref: "streamsim/chiller" } });
  });

  it("adds equipment, saves the file with its version and composes the domain over the one it read", async () => {
    const { store, fake } = await opened();
    await store.add({ use: "streamsim/pump@1", preset: "small", name: "P-1", area: "Site" });
    expect(store.snapshot()).toMatchObject({ selected: "P-1" });
    expect((await store.presets("streamsim/pump@1")).preferred).toBe("small");
    await store.save();
    expect(fake.writes[0]).toMatchObject({ path, etag: "c1" });
    expect(fake.calls.find((call) => call.name === "domain.compose")?.args).toEqual({
      composition: path,
      out: "domains/lab.domain.json",
      etag: "d1",
    });
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.saved).toBe(state.doc);
    expect(state.outcome).toEqual({ kind: "composed", summary: "Composed 3 channels, 2 faults, 1 effectors." });
    expect(equipmentOf(state.doc).map((entry) => entry.names[0])).toEqual(["CH-1", "P-1"]);
  });

  it("puts the binary's issues on the saved instances and keeps a refusal without issues", async () => {
    const fake = bridge();
    fake.actions.set(
      "domain.compose",
      failed({ ...failure("composition_invalid", 422), issues: [{ pointer: "/equipment/0/use", message: "unknown" }] }),
    );
    const { store } = await opened(fake);
    await store.save();
    expect(store.snapshot()).toMatchObject({ outcome: { kind: "refused" }, failure: null, saving: false });
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.issues?.instances.get("CH-1")).toEqual(["unknown"]);
    fake.writeOutcomes.set(path, failed(failure("conflict", 409, "changed on disk")));
    await store.save();
    expect(store.snapshot()).toMatchObject({ failure: { code: "conflict" }, saving: false });
  });

  it("creates a new composition from a base and refuses an unreadable file", async () => {
    const fake = bridge();
    const store = new EditorStore(fake);
    expect(await store.create(path, "lab-two", "Lab two")).toBe("compositions/lab-two.composition.json");
    expect(fake.writes[0]).toMatchObject({ path: "compositions/lab-two.composition.json", etag: "" });
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toMatchObject({ id: "lab-two", equipment: [] });
    fake.files.set("compositions/bad.composition.json", success({ path: "x", content: "[", etag: "e" }));
    expect(await store.create("compositions/bad.composition.json", "x-y-z", "X")).toMatchObject({
      code: "composition_unreadable",
    });
    await store.open("compositions/bad.composition.json");
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { code: "composition_unreadable" } });
    store.select("CH-1");
    expect(store.snapshot()).toMatchObject({ phase: "failed" });
  });

  it("recovers from a conflict by reloading, keeps a dirty draft and refreshes a clean one", async () => {
    const { store, fake } = await opened();
    const changed = { ...doc, title: "Lab changed" };
    fake.files.set(path, success({ path, content: JSON.stringify(changed), etag: "c2" }));
    await store.open(path);
    expect(store.snapshot()).toMatchObject({ etag: "c2", doc: { title: "Lab changed" } });
    store.edit((current) => ({ ...current, title: "Mine" }));
    fake.files.set(path, success({ path, content: JSON.stringify(doc), etag: "c3" }));
    await store.open(path);
    expect(store.snapshot()).toMatchObject({ etag: "c2", doc: { title: "Mine" } });
    fake.writeOutcomes.set(path, failed(failure("conflict", 409, "changed on disk")));
    await store.save();
    expect(store.snapshot()).toMatchObject({ failure: { code: "conflict" } });
    await store.reload();
    expect(store.snapshot()).toMatchObject({ etag: "c3", failure: null, doc: { title: "Lab" } });
  });

  it("composes over the domain version read when the plant was opened and says why compose failed", async () => {
    const { store, fake } = await opened();
    fake.files.set("domains/lab.domain.json", success({ path: "x", content: "{}", etag: "d9" }));
    fake.actions.set("domain.compose", failed(failure("conflict", 409, "the domain changed on disk")));
    await store.save();
    expect(fake.calls.find((call) => call.name === "domain.compose")?.args).toMatchObject({ etag: "d1" });
    expect(store.snapshot()).toMatchObject({
      failure: null,
      outcome: { kind: "refused", message: "the domain changed on disk" },
    });
    store.edit((current) => current, null);
    expect(store.snapshot()).toMatchObject({ selected: null });
  });

  it("refuses a new plant whose domain already exists", async () => {
    const fake = bridge();
    const store = new EditorStore(fake);
    expect(await store.create(path, "lab", "Lab again")).toMatchObject({ code: "plant_exists" });
    expect(fake.writes).toEqual([]);
  });

  it("undoes and redoes every edit, revert included, and forgets the redo after a new edit", async () => {
    const { store } = await opened();
    const title = () => {
      const state = store.snapshot();
      return state.phase === "ready" ? state.doc.title : "";
    };
    store.edit((doc) => ({ ...doc, title: "One" }));
    store.edit((doc) => ({ ...doc, title: "Two" }));
    store.revert();
    expect(title()).toBe("Lab");
    store.undo();
    expect(title()).toBe("Two");
    store.undo();
    store.undo();
    store.undo();
    expect(title()).toBe("Lab");
    store.redo();
    expect(title()).toBe("One");
    store.edit((doc) => ({ ...doc, title: "Three" }));
    store.redo();
    expect(title()).toBe("Three");
    expect(store.snapshot()).toMatchObject({ future: [] });
    const before = store.snapshot();
    store.edit((doc) => doc);
    expect(store.snapshot()).toBe(before);
  });
});
