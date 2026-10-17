import { describe, expect, it, vi } from "vitest";
import { failure as failed, success } from "../../ports";
import { flushPromises } from "../../../test/deferred";
import { equipmentOf } from "../../../model/editor";
import { FakeBridge, cannedFailure, cannedFile } from "../../../test/fakeBridge";
import { canSave, emptyPlant } from "./editorSelectors";
import { createEditorSession } from "./editorSession";

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
  fake.addFile(path, doc, "c1");
  fake.addFile("domains/lab.domain.json", "{}", "d1");
  fake.actions.set("component.list", success({ components: [{ id: "streamsim/pump", kind: "equipment" }] }));
  fake.actions.set("component.show", success({ document: { presets: { small: {} }, default_preset: "small" } }));
  fake.actions.set("domain.compose", success({ channels: 3, faults: 2, effectors: 1 }));
  return fake;
}

async function opened(fake = bridge()) {
  const store = createEditorSession(fake.editor);
  await store.open(path);
  const state = store.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return { store, fake, state };
}

describe("editor store", () => {
  it("keeps Save off for an empty plant that has never been checked", async () => {
    const empty = { ...doc, equipment: [], areas: [] };
    const fake = bridge();
    fake.addFile(path, empty, "c1");
    const { store, state } = await opened(fake);
    expect(emptyPlant(state)).toBe(true);
    expect(canSave(state)).toBe(false);
    store.edit((current) => ({ ...current, title: "Changed" }));
    const edited = store.snapshot();
    if (edited.phase !== "ready") throw new Error(edited.phase);
    expect(canSave(edited)).toBe(true);
  });

  it("writes each save over the version the last write returned and composes over the domain it read", async () => {
    const { store, fake } = await opened();
    fake.writeOutcomes.set(path, success({ path, etag: "c2" }));
    store.edit((current) => ({ ...current, title: "One" }));
    await store.save();
    store.edit((current) => ({ ...current, title: "Two" }));
    await store.save();
    expect(fake.writes.map((write) => write.etag)).toEqual(["c1", "c2"]);
    const composes = fake.calls.filter((call) => call.name === "domain.compose");
    expect(composes.map((call) => (call.args as { etag: string }).etag)).toEqual(["d1", "d1"]);
  });

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
    expect(state.save).toEqual({
      kind: "composed",
      summary: "Composed 3 channels, 2 faults, 1 effectors.",
      doc: state.doc,
    });
    expect(equipmentOf(state.doc).map((entry) => entry.names[0])).toEqual(["CH-1", "P-1"]);
  });

  it("puts the binary's issues on the saved instances and keeps a refusal without issues", async () => {
    const fake = bridge();
    fake.actions.set(
      "domain.compose",
      failed({
        ...cannedFailure("composition_invalid", 422),
        issues: [{ pointer: "/equipment/0/use", message: "unknown" }],
      }),
    );
    const { store } = await opened(fake);
    await store.save();
    expect(store.snapshot()).toMatchObject({ save: { kind: "compose-failed" } });
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.issues?.instances.get("CH-1")).toEqual(["unknown"]);
    fake.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    await store.save();
    expect(store.snapshot()).toMatchObject({ save: { kind: "write-failed", failure: { kind: "conflict" } } });
  });

  it("refuses to open a file that is not a composition", async () => {
    const fake = bridge();
    const store = createEditorSession(fake.editor);
    fake.addFile("compositions/bad.composition.json", "[");
    await store.open("compositions/bad.composition.json");
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      failure: { kind: "unreadable", diagnostic: { displayCode: "composition_unreadable" } },
    });
    store.select("CH-1");
    expect(store.snapshot()).toMatchObject({ phase: "failed" });
  });

  it("refreshes a clean draft from disk and recovers from a conflict by reloading", async () => {
    const { store, fake } = await opened();
    fake.persistWrites();
    const changed = { ...doc, title: "Lab changed" };
    fake.addFile(path, changed, "c2");
    await store.reload();
    expect(store.snapshot()).toMatchObject({ doc: { title: "Lab changed" } });
    store.edit((current) => ({ ...current, title: "Mine" }));
    fake.addFile(path, doc, "c3");
    await store.save();
    expect(store.snapshot()).toMatchObject({ save: { kind: "write-failed", failure: { kind: "conflict" } } });
    await store.reload();
    expect(store.snapshot()).toMatchObject({ save: { kind: "idle" }, doc: { title: "Lab" } });
    store.edit((current) => ({ ...current, title: "Again" }));
    await store.save();
    expect(fake.writes.at(-1)).toMatchObject({ etag: "c3" });
  });

  it("composes over the domain version read when the plant was opened and says why compose failed", async () => {
    const { store, fake } = await opened();
    fake.addFile("domains/lab.domain.json", "{}", "d9");
    fake.actions.set("domain.compose", failed(cannedFailure("conflict", 409, "the domain changed on disk")));
    await store.save();
    expect(fake.calls.find((call) => call.name === "domain.compose")?.args).toMatchObject({ etag: "d1" });
    expect(store.snapshot()).toMatchObject({
      save: { kind: "compose-failed", failure: { diagnostic: { message: "the domain changed on disk" } } },
    });
    store.edit((current) => current, null);
    expect(store.snapshot()).toMatchObject({ selected: null });
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

  it("keeps the edit on screen when a discard arrives while its save is in flight", async () => {
    const { store, fake } = await opened();
    const write = fake.holdWrite();
    store.edit((current) => ({ ...current, title: "Mine" }));
    const saving = store.save();
    await flushPromises();
    store.revert();
    expect(store.snapshot()).toMatchObject({ save: { kind: "saving" }, doc: { title: "Mine" } });
    write[0]?.resolve(success({ path, etag: "c2" }));
    await saving;
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.saved).toBe(state.doc);
    expect(state.doc.title).toBe("Mine");
  });

  it("drops a selection that undo or redo takes out of the plant", async () => {
    const { store } = await opened();
    store.edit(
      (doc) => ({ ...doc, equipment: [...(doc.equipment as unknown[]), { use: "x@1", names: ["P-9"] }] }),
      "P-9",
    );
    store.undo();
    expect(store.snapshot()).toMatchObject({ selected: null });
    store.select("CH-1");
    store.redo();
    expect(store.snapshot()).toMatchObject({ selected: "CH-1" });
  });

  it("ignores an earlier visit's answer when the same composition is opened again", async () => {
    const fake = bridge();
    const other = "compositions/other.composition.json";
    fake.addFile(other, { ...doc, id: "other" }, "c2");
    const held = fake.holdRead(path);
    const store = createEditorSession(fake.editor);
    const first = store.open(path);
    await store.open(other);
    const third = store.open(path);
    held[0]?.resolve(cannedFile(path, { ...doc, title: "stale" }, "c1"));
    await first;
    const stale = store.snapshot();
    expect(stale).not.toMatchObject({ phase: "ready", doc: { title: "stale" } });
    held[1]?.resolve(cannedFile(path, doc, "c1"));
    await third;
    expect(store.snapshot()).toMatchObject({ phase: "ready", path, doc: { title: "Lab" } });
  });

  it("rejects a late open after close and accepts the reopened document", async () => {
    const fake = bridge();
    const reads = fake.holdRead(path);
    const store = createEditorSession(fake.editor);
    const first = store.open(path);
    store.close();
    const reopened = store.open(path);
    reads[0]?.resolve(cannedFile(path, { ...doc, title: "late" }, "c1"));
    await first;
    expect(store.snapshot()).toMatchObject({ phase: "loading", path });
    reads[1]?.resolve(cannedFile(path, doc, "c1"));
    await reopened;
    expect(store.snapshot()).toMatchObject({ phase: "ready", doc: { title: "Lab" } });
  });

  it("does not apply a late compose version to a different reopened composition", async () => {
    const fake = bridge();
    const other = "compositions/other.composition.json";
    const otherDomain = "domains/other.domain.json";
    fake.addFile(other, { ...doc, id: "other" }, "c2");
    fake.addFile(otherDomain, "{}", "d2");
    const { store } = await opened(fake);
    const heldVersion = fake.holdRead("domains/lab.domain.json");
    const saving = store.save();
    await vi.waitFor(() => {
      expect(heldVersion).toHaveLength(1);
    });
    store.close();
    await store.open(other);
    heldVersion[0]?.resolve(cannedFile("domains/lab.domain.json", "{}", "d9"));
    await saving;
    expect(store.snapshot()).toMatchObject({
      phase: "ready",
      path: other,
      domain: otherDomain,
      save: { kind: "idle" },
    });
  });
});
