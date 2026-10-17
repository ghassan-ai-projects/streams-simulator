import { describe, expect, it } from "vitest";
import { changeMetric, renameKind } from "../../../model/devices";
import { failure as failed, success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Draft } from "./deviceDraft";
import { canSave, isBusy, isDirty } from "./deviceSelectors";
import { createDeviceSession } from "./deviceStore";

const path = "kinds/fan.kind.json";
const fan = {
  kind: "fan",
  metrics: [{ name: "duty", data_type: "number" }],
  commands: [{ name: "stop", args_schema: { type: "object", properties: {} } }],
};
const valid = success({
  valid: true,
  kinds: [{ path, kind: "fan", kind_digest: "sha256:1", metrics: 1, commands: 1 }],
});

function bridge() {
  const fake = new FakeBridge();
  fake.addFile(path, fan, "k1");
  fake.actions.set("kind.validate", valid);
  return fake;
}

function ready(store: ReturnType<typeof createDeviceSession>): Draft {
  const state = store.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function opened(fake = bridge()) {
  const store = createDeviceSession(fake.devices);
  await store.open(path);
  return { store, fake };
}

function refusal(pointer: string, message: string) {
  return failed({
    ...cannedFailure("command_failed", 422, "kind file fails its schema"),
    issues: [{ pointer, message }],
  });
}

describe("device session", () => {
  it("reads the file and asks the binary to check it", async () => {
    const { store, fake } = await opened();
    const draft = ready(store);
    expect(draft.selected).toEqual({ part: "kind" });
    expect(draft.check).toMatchObject({ kind: "valid", summary: { digest: "sha256:1", metrics: 1, commands: 1 } });
    expect(fake.calls).toEqual([{ name: "kind.validate", args: { path } }]);
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

  it("writes the device over the version it read, then checks the saved file", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "%" }));
    expect(canSave(ready(store))).toBe(true);
    await store.save();
    expect(fake.writes[0]).toMatchObject({ path, etag: "k1" });
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toMatchObject({ metrics: [{ name: "duty", unit: "%" }] });
    expect(fake.calls.filter((call) => call.name === "kind.validate")).toHaveLength(2);
    const draft = ready(store);
    expect(draft.saved).toBe(draft.doc);
    expect(draft.save).toEqual({ kind: "idle" });
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Cel" }));
    await store.save();
    expect(fake.writes.map((write) => write.etag)).toEqual(["k1", "written"]);
  });

  it("puts each issue of the binary on the reading it names", async () => {
    const { store, fake } = await opened();
    fake.actions.set("kind.validate", refusal("/metrics/0/data_type", "not one of the types"));
    store.edit((kind) => changeMetric(kind, "m1", { dataType: "nope" }));
    await store.save();
    const draft = ready(store);
    expect(draft.check).toMatchObject({ kind: "refused" });
    expect(draft.issues).toEqual([{ part: "metric", key: "m1", field: "data_type", message: "not one of the types" }]);
    fake.actions.set("kind.validate", valid);
    await store.check();
    expect(ready(store).issues).toEqual([]);
  });

  it("keeps what could not be checked apart from what the binary refused", async () => {
    const { store, fake } = await opened();
    fake.actions.set("kind.validate", failed(cannedFailure("bridge_failed", 500, "down")));
    await store.check();
    expect(ready(store).check).toMatchObject({ kind: "failed", failure: { kind: "unavailable" } });
    expect(isBusy(ready(store))).toBe(false);
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
    store.edit((kind) => changeMetric(kind, "m1", { unit: "%" }));
    fake.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    await store.save();
    expect(ready(store).save).toMatchObject({ kind: "write-failed", failure: { kind: "conflict" } });
    expect(canSave(ready(store))).toBe(false);
    fake.writeOutcomes.set(path, failed(cannedFailure("bridge_failed", 500, "disk full")));
    await store.reload();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    await store.save();
    expect(canSave(ready(store))).toBe(true);
    fake.writeOutcomes.delete(path);
    await store.save();
    expect(ready(store).save).toEqual({ kind: "idle" });
  });

  it("reloads the file as it is on disk and drops the draft", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { unit: "%" }));
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
    expect(fake.calls).toHaveLength(2);
    fake.handlers.set("kind.validate", () => valid);
    await store.reload();
    slow.resolve(refusal("/kind", "late"));
    await flushPromises();
    expect(ready(store).check.kind).toBe("valid");
    expect(ready(store).issues).toEqual([]);
    store.close();
    await store.check();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });
});
