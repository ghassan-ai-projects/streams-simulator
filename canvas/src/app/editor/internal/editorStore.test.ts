import { describe, expect, it } from "vitest";
import type { FileContent } from "../../ports";
import { failure as failed, success } from "../../ports";
import type { Deferred } from "../../../test/deferred";
import { deferred, flushPromises } from "../../../test/deferred";
import { equipmentOf } from "../../../model/editor";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { canSave, emptyPlant } from "./editorStore";
import { editorStore } from "../../../test/stores";

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
  const store = editorStore(fake.editor);
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
    expect(state.outcome).toEqual({
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
    expect(store.snapshot()).toMatchObject({ outcome: { kind: "refused" }, failure: null, saving: false });
    const state = store.snapshot();
    if (state.phase !== "ready") throw new Error(state.phase);
    expect(state.issues?.instances.get("CH-1")).toEqual(["unknown"]);
    fake.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    await store.save();
    expect(store.snapshot()).toMatchObject({ failure: { kind: "conflict" }, saving: false });
  });

  it("refuses to open a file that is not a composition", async () => {
    const fake = bridge();
    const store = editorStore(fake.editor);
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
    const changed = { ...doc, title: "Lab changed" };
    fake.addFile(path, changed, "c2");
    await store.reload();
    expect(store.snapshot()).toMatchObject({ etag: "c2", doc: { title: "Lab changed" } });
    store.edit((current) => ({ ...current, title: "Mine" }));
    fake.addFile(path, doc, "c3");
    fake.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    await store.save();
    expect(store.snapshot()).toMatchObject({ failure: { kind: "conflict" } });
    await store.reload();
    expect(store.snapshot()).toMatchObject({ etag: "c3", failure: null, doc: { title: "Lab" } });
  });

  it("composes over the domain version read when the plant was opened and says why compose failed", async () => {
    const { store, fake } = await opened();
    fake.addFile("domains/lab.domain.json", "{}", "d9");
    fake.actions.set("domain.compose", failed(cannedFailure("conflict", 409, "the domain changed on disk")));
    await store.save();
    expect(fake.calls.find((call) => call.name === "domain.compose")?.args).toMatchObject({ etag: "d1" });
    expect(store.snapshot()).toMatchObject({
      failure: null,
      outcome: { kind: "refused", message: "the domain changed on disk" },
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
    const write = deferred<Canned<{ path: string; etag: string }>>();
    fake.writeFile = () => write.promise;
    store.edit((current) => ({ ...current, title: "Mine" }));
    const saving = store.save();
    await flushPromises();
    store.revert();
    expect(store.snapshot()).toMatchObject({ saving: true, doc: { title: "Mine" } });
    write.resolve(success({ path, etag: "c2" }));
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
    const held: Deferred<Canned<FileContent>>[] = [];
    const read = fake.readFile.bind(fake);
    fake.readFile = (requested: string) => {
      if (requested !== path) return read(requested);
      const gate = deferred<Canned<FileContent>>();
      held.push(gate);
      return gate.promise;
    };
    const store = editorStore(fake.editor);
    const first = store.open(path);
    await store.open(other);
    const third = store.open(path);
    held[0]?.resolve(success({ path, content: JSON.stringify({ ...doc, title: "stale" }), etag: "c1" }));
    await first;
    const stale = store.snapshot();
    expect(stale).not.toMatchObject({ phase: "ready", doc: { title: "stale" } });
    held[1]?.resolve(success({ path, content: JSON.stringify(doc), etag: "c1" }));
    await third;
    expect(store.snapshot()).toMatchObject({ phase: "ready", path, doc: { title: "Lab" } });
  });

  it("rejects a late open after close and accepts the reopened document", async () => {
    const fake = bridge();
    const reads: Deferred<Canned<FileContent>>[] = [];
    const read = fake.readFile.bind(fake);
    fake.readFile = (requested: string) => {
      if (requested !== path) return read(requested);
      const gate = deferred<Canned<FileContent>>();
      reads.push(gate);
      return gate.promise;
    };
    const store = editorStore(fake.editor);
    const first = store.open(path);
    store.close();
    const reopened = store.open(path);
    reads[0]?.resolve(success({ path, content: JSON.stringify({ ...doc, title: "late" }), etag: "c1" }));
    await first;
    expect(store.snapshot()).toMatchObject({ phase: "loading", path });
    reads[1]?.resolve(success({ path, content: JSON.stringify(doc), etag: "c1" }));
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
    const heldVersion = deferred<Canned<FileContent>>();
    const versionRead = deferred<true>();
    const read = fake.readFile.bind(fake);
    fake.readFile = (requested) => {
      if (requested === "domains/lab.domain.json") {
        versionRead.resolve(true);
        return heldVersion.promise;
      }
      return read(requested);
    };
    const saving = store.save();
    await versionRead.promise;
    store.close();
    await store.open(other);
    heldVersion.resolve(success({ path: "domains/lab.domain.json", content: "{}", etag: "d9" }));
    await saving;
    expect(store.snapshot()).toMatchObject({ phase: "ready", path: other, domainEtag: "d2", outcome: null });
  });
});
