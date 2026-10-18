import { describe, expect, it, vi } from "vitest";
import { changeMetric, renameKind } from "../../../model/devices";
import { failure as failed, success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import { bridge, checks, fan, opened, path, ready, refusal, valid } from "../../../test/deviceSessionHarness";
import { kindCheck } from "../../../test/kindCheckFake";
import { createDeviceSession } from "./deviceStore";
import { canSave, isBusy, isDirty } from "./deviceSelectors";

describe("device session", () => {
  it("reads the file, asks the binary to check it, and asks what a twin would need of the draft", async () => {
    const { store, fake } = await opened();
    const draft = ready(store);
    expect(draft.selected).toEqual({ part: "kind" });
    expect(draft.check).toMatchObject({ kind: "valid", summary: { digest: "sha256:1", metrics: 1, commands: 1 } });
    expect(fake.calls.filter((call) => call.name === "kind.validate")).toEqual([
      { name: "kind.validate", args: { path } },
    ]);
    expect(checks(fake, true)).toHaveLength(1);
    expect(JSON.parse(String(checks(fake, true)[0]?.args.content))).toMatchObject({ kind: "fan" });
    expect(draft.review).toEqual({ state: "done", schema: [], needs: [] });
    expect(isDirty(draft)).toBe(false);
    expect(canSave(draft)).toBe(false);
  });

  it("waits while the file is read and says why it cannot open one", async () => {
    const fake = bridge();
    const gates = fake.holdRead(path, 1);
    const store = createDeviceSession(fake.devices);
    const opening = store.open(path);
    expect(store.snapshot()).toEqual({ phase: "loading", path });
    await flushPromises();
    gates[0]?.resolve(success({ path, content: JSON.stringify(fan), etag: "k1" }));
    await opening;
    expect(ready(store).path).toBe(path);
    fake.files.delete(path);
    await store.open(path);
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { kind: "missing" } });
    fake.addFile(path, "not json");
    await store.open(path);
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { kind: "unreadable" } });
  });

  it("tracks edits, and a revert puts the saved device back", async () => {
    const { store } = await opened();
    store.edit((kind) => renameKind(kind, "fan"));
    expect(isDirty(ready(store))).toBe(false);
    store.addMetric();
    store.addCommand();
    expect(ready(store).selected).toEqual({ part: "command", key: "c2" });
    expect(isDirty(ready(store))).toBe(true);
    store.revert();
    expect(isDirty(ready(store))).toBe(false);
    expect(ready(store).selected).toEqual({ part: "kind" });
    store.revert();
    expect(ready(store).doc).toBe(ready(store).saved);
  });

  it("removes the selected item and selects the device again", async () => {
    const { store } = await opened();
    store.select({ part: "metric", key: "m1" });
    store.remove({ part: "metric", key: "m1" });
    expect(ready(store).doc.metrics).toEqual([]);
    expect(ready(store).selected).toEqual({ part: "kind" });
  });
});

describe("saving", () => {
  it("asks the binary about the draft before anything is written, then writes it over the version it read", async () => {
    const { store, fake } = await opened();
    let written = -1;
    fake.handlers.set("kind.check", (args) => {
      if (args.twin === false) written = fake.writes.length;
      return kindCheck(args);
    });
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    expect(canSave(ready(store))).toBe(true);
    await store.save();
    expect(written).toBe(0);
    expect(fake.writes[0]).toMatchObject({ path, etag: "k1" });
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toMatchObject({ metrics: [{ name: "duty", unit: "Pa" }] });
    const draft = ready(store);
    expect(draft.saved).toBe(draft.doc);
    expect(draft.save).toEqual({ kind: "idle" });
    expect(draft.check).toMatchObject({ kind: "valid", summary: { digest: "sha256:abc" } });
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Cel" }));
    await store.save();
    expect(fake.writes.map((write) => write.etag)).toEqual(["k1", "written"]);
  });

  it("leaves a refused definition out of the workspace and puts each issue on the reading it names", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { name: "bad name" }));
    await store.save();
    expect(fake.writes).toEqual([]);
    const draft = ready(store);
    expect(draft.save).toMatchObject({ kind: "refused", failure: { kind: "refused" } });
    expect(draft.review.schema.map((issue) => `${issue.key}:${issue.field}`)).toEqual(["m1:name"]);
    expect(isDirty(draft)).toBe(true);
    store.edit((kind) => changeMetric(kind, "m1", { name: "speed" }));
    expect(ready(store).save).toEqual({ kind: "idle" });
    await store.save();
    expect(fake.writes).toHaveLength(1);
    await vi.waitFor(() => {
      expect(ready(store).review.schema).toEqual([]);
    });
  });

  it("keeps a write failure apart from a refusal, and writes nothing when the check could not run", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    fake.handlers.set("kind.check", () => failed(cannedFailure("bridge_failed", 500, "down")));
    await store.save();
    expect(ready(store).save).toMatchObject({ kind: "write-failed", failure: { kind: "unavailable" } });
    expect(fake.writes).toEqual([]);
  });

  it("does not save a device with a blank name", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { name: " " }));
    expect(canSave(ready(store))).toBe(false);
    await store.save();
    expect(fake.writes).toEqual([]);
  });

  it("stops at a conflict until the file is reloaded, and offers a retry after any other failure", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    fake.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    await store.save();
    expect(ready(store).save).toMatchObject({ kind: "write-failed", failure: { kind: "conflict" } });
    expect(canSave(ready(store))).toBe(false);
    fake.writeOutcomes.set(path, failed(cannedFailure("bridge_failed", 500, "disk full")));
    await store.reload();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Cel" }));
    await store.save();
    expect(canSave(ready(store))).toBe(true);
    fake.writeOutcomes.delete(path);
    await store.save();
    expect(ready(store).save).toEqual({ kind: "idle" });
  });

  it("ignores a save that finishes after the session closed", async () => {
    const { store, fake } = await opened();
    const gates = fake.holdWrite(1);
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    const saving = store.save();
    await flushPromises();
    expect(isBusy(ready(store))).toBe(true);
    store.close();
    gates[0]?.resolve(success({ path, etag: "late" }));
    await saving;
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });
});

describe("checking the saved file", () => {
  it("asks the binary about the file on disk and says what it could not check apart from what it refused", async () => {
    const { store, fake } = await opened();
    fake.actions.set("kind.validate", refusal("/kind", "refused"));
    await store.check();
    expect(ready(store).check).toMatchObject({ kind: "refused" });
    fake.actions.set("kind.validate", failed(cannedFailure("bridge_failed", 500, "down")));
    await store.check();
    expect(ready(store).check).toMatchObject({ kind: "failed", failure: { kind: "unavailable" } });
    expect(isBusy(ready(store))).toBe(false);
  });

  it("reloads the file as it is on disk, drops the draft and the behaviour, and does nothing once closed", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    fake.addFile(path, { ...fan, kind: "fan2" }, "k2");
    await store.reload();
    expect(ready(store).doc.kind).toBe("fan2");
    store.close();
    expect(store.snapshot()).toEqual({ phase: "idle" });
    await store.reload();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });

  it("ignores a check that a newer one overtook, and any check after the session closed", async () => {
    const { store, fake } = await opened();
    const slow = deferred<ReturnType<typeof refusal>>();
    fake.handlers.set("kind.validate", () => slow.promise);
    void store.check();
    await flushPromises();
    expect(isBusy(ready(store))).toBe(true);
    await store.check();
    expect(fake.calls.filter((call) => call.name === "kind.validate")).toHaveLength(2);
    fake.handlers.set("kind.validate", () => valid);
    await store.reload();
    slow.resolve(refusal("/kind", "late"));
    await flushPromises();
    expect(ready(store).check.kind).toBe("valid");
    store.close();
    await store.check();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });
});
