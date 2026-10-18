import { describe, expect, it, vi } from "vitest";
import { chooseInstrument, renameTwin, startAt } from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import { flushPromises } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import { failure } from "../../../test/outcome";
import { choicesPath, cooler, coolerPath, kindOf, sensorPath, thermistor, twinBridge } from "../../../test/twinFixture";
import type { TwinDraft } from "./twinDraft";
import { canCreate, hasUnsavedChoices, isChoicesDirty, othersOf, pairsOf } from "./twinSelectors";
import { createTwinSession } from "./twinStore";

function ready(twin: ReturnType<typeof createTwinSession>): TwinDraft {
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function opened(fake: FakeBridge = twinBridge(), lead: () => KindFile | null = () => cooler.kind) {
  const twin = createTwinSession(fake.devices, lead);
  await twin.open(coolerPath);
  return { twin, fake };
}

function calls(fake: FakeBridge, name: string) {
  return fake.calls.filter((call) => call.name === name);
}

describe("opening the behaviour of a device", () => {
  it("reads the choices file, the device list, the binary's suggestion and the instruments for each unit", async () => {
    const { twin, fake } = await opened();
    const draft = ready(twin);
    expect(draft.choicesPath).toBe(choicesPath);
    expect(draft.members.map((member) => member.kind.kind)).toEqual(["cooler"]);
    expect(draft.touched).toBe(false);
    expect(draft.edited).toBe(false);
    expect(draft.saved).toBe(draft.doc);
    expect(draft.doc).toMatchObject({ id: "cooler-sensor-twin", title: "Twin of cooler, sensor" });
    expect(draft.doc.kinds[0]?.metrics.m1?.instrument).toBe("none");
    expect(Object.keys(draft.instruments)).toEqual(["Cel", "%"]);
    expect(draft.instruments.Cel?.map((instrument) => instrument.match)).toEqual(["convertible"]);
    expect(calls(fake, "component.list").map((call) => call.args)).toEqual([{ unit: "Cel" }, { unit: "%" }]);
    expect(calls(fake, "kind.suggest")).toHaveLength(1);
    expect(othersOf(draft).map((entry) => entry.path)).toEqual([sensorPath]);
    expect(fake.reads).toContain(choicesPath);
    expect(pairsOf(draft)).toHaveLength(1);
    expect(isChoicesDirty(draft)).toBe(false);
    expect(draft.onDisk).toBe(false);
  });

  it("reopens with the choices saved before, with the other devices they name", async () => {
    const fake = twinBridge();
    const first = await opened(fake);
    await first.twin.include(sensorPath);
    first.twin.edit((choices) => chooseInstrument(choices, coolerPath, "m1", thermistor.ref));
    first.twin.edit((choices) => startAt(choices, coolerPath, "m1", 20));
    await first.twin.persist();
    const draft = ready((await opened(fake)).twin);
    expect(draft.members.map((member) => member.kind.kind)).toEqual(["cooler", "sensor"]);
    expect(draft.touched).toBe(true);
    expect(draft.doc.kinds[0]?.metrics.m1).toEqual({ instrument: thermistor.ref, initial: 20 });
    expect(isChoicesDirty(draft)).toBe(false);
    expect(othersOf(draft)).toEqual([]);
  });

  it("adds what the interface gained since the choices were saved, which is then waiting to be saved", async () => {
    const fake = twinBridge();
    const first = await opened(fake);
    first.twin.edit((choices) => renameTwin(choices, "kept-twin"));
    await first.twin.persist();
    const metric = { name: "pressure", data_type: "number", unit: "Pa", resolution: 1 };
    const grown = { kind: "cooler", metrics: [metric], commands: [] };
    const named = kindOf(grown);
    const draft = ready((await opened(fake, () => named)).twin);
    expect(draft.doc.id).toBe("kept-twin");
    expect(Object.keys(draft.doc.kinds[0]?.metrics ?? {})).toEqual(["m1"]);
    expect(isChoicesDirty(draft)).toBe(true);
    expect(hasUnsavedChoices(draft)).toBe(false);
  });

  it("opens without the library or the device list, and says the instruments could not be read", async () => {
    const fake = twinBridge();
    fake.handlers.set("component.list", () => failure(cannedFailure("boom", 500)));
    fake.workspaceOutcome = failure(cannedFailure("boom", 500));
    const draft = ready((await opened(fake)).twin);
    expect(draft.instruments).toEqual({ Cel: null, "%": null });
    expect(draft.known).toEqual([]);
    expect(draft.doc.kinds[0]?.metrics.m1?.instrument).toBe("none");
  });

  it("goes on with its saved choices when the binary has no suggestion, and says why when it has no choices either", async () => {
    const fake = twinBridge();
    const first = await opened(fake);
    first.twin.edit((choices) => renameTwin(choices, "kept-twin"));
    await first.twin.persist();
    fake.handlers.delete("kind.suggest");
    const draft = ready((await opened(fake)).twin);
    expect(draft.doc.id).toBe("kept-twin");
    expect(draft.suggested).toBeNull();
    fake.files.delete(choicesPath);
    const { twin } = await opened(fake);
    expect(twin.snapshot()).toMatchObject({ phase: "failed", failure: { kind: "unreadable" } });
    fake.handlers.set("kind.suggest", () => failure(cannedFailure("kind_failed", 422, "refused")));
    expect((await opened(fake)).twin.snapshot()).toMatchObject({ phase: "failed", failure: { kind: "refused" } });
  });

  it("starts from the suggestion when the saved choices file is not a document", async () => {
    const fake = twinBridge();
    fake.addFile(choicesPath, "not json", "c1");
    const draft = ready((await opened(fake)).twin);
    expect(draft.doc.id).toBe("cooler-sensor-twin");
    expect(draft.touched).toBe(false);
  });

  it.each([
    [
      "the choices file cannot be read",
      (fake: FakeBridge) => fake.files.set(choicesPath, failure(cannedFailure("boom", 500))),
      "unavailable",
    ],
  ])("says why it cannot open when %s", async (_name, spoil, kind) => {
    const fake = twinBridge();
    spoil(fake);
    const { twin } = await opened(fake);
    expect(twin.snapshot()).toMatchObject({ phase: "failed", path: coolerPath, failure: { kind } });
  });

  it("says why it cannot open when the device is not open", async () => {
    const { twin } = await opened(twinBridge(), () => null);
    expect(twin.snapshot()).toMatchObject({
      phase: "failed",
      failure: { diagnostic: { message: "The device is not open." } },
    });
  });

  it("ignores an answer that arrives after it was closed", async () => {
    const fake = twinBridge();
    const gates = fake.holdRead(choicesPath, 1);
    const twin = createTwinSession(fake.devices, () => cooler.kind);
    const opening = twin.open(coolerPath);
    expect(twin.snapshot()).toEqual({ phase: "loading", path: coolerPath });
    await flushPromises();
    twin.close();
    gates[0]?.resolve(failure(cannedFailure("not_found", 404)));
    await opening;
    expect(twin.snapshot()).toEqual({ phase: "idle" });
    await twin.reload();
    expect(twin.snapshot()).toEqual({ phase: "idle" });
  });
});

describe("saving the choices", () => {
  it("writes only what changed, one write at a time, each over the version the last one made", async () => {
    const { twin, fake } = await opened();
    await twin.persist();
    expect(fake.writes).toHaveLength(0);
    twin.edit((choices) => renameTwin(choices, "plant-a"));
    expect(hasUnsavedChoices(ready(twin))).toBe(true);
    const first = twin.persist();
    twin.edit((choices) => renameTwin(choices, "plant-b"));
    const second = twin.persist();
    await Promise.all([first, second]);
    expect(fake.writes.map((write) => write.etag)).toEqual(["", "e1"]);
    expect(JSON.parse(fake.writes[1]?.content ?? "{}")).toMatchObject({ id: "plant-b", version: "0.1.0" });
    expect(ready(twin).save).toEqual({ kind: "idle" });
    expect(isChoicesDirty(ready(twin))).toBe(false);
    expect(hasUnsavedChoices(ready(twin))).toBe(false);
  });

  it("says when the file changed on disk and does not write over it", async () => {
    const { twin, fake } = await opened();
    fake.writeOutcomes.set(choicesPath, failure(cannedFailure("conflict", 409, "changed")));
    twin.edit((choices) => renameTwin(choices, "plant-a"));
    await twin.persist();
    expect(ready(twin).save).toMatchObject({ kind: "write-failed", failure: { kind: "conflict" } });
    expect(hasUnsavedChoices(ready(twin))).toBe(true);
    expect(canCreate(ready(twin), { phase: "idle" })).toBe(false);
    await twin.create();
    expect(fake.calls.some((call) => call.name === "kind.twin")).toBe(false);
    fake.writeOutcomes.clear();
    await twin.reload();
    expect(ready(twin).save).toEqual({ kind: "idle" });
  });

  it("writes unsaved choices when the screen is left", async () => {
    const { twin, fake } = await opened();
    twin.edit((choices) => renameTwin(choices, "plant-a"));
    twin.close();
    await vi.waitFor(() => {
      expect(fake.writes).toHaveLength(1);
    });
    expect(twin.snapshot()).toEqual({ phase: "idle" });
    twin.close();
    expect(fake.writes).toHaveLength(1);
  });

  it("keeps the state when an edit changes nothing, and drops edits on request", async () => {
    const { twin } = await opened();
    const before = ready(twin);
    twin.edit((choices) => choices);
    expect(ready(twin)).toBe(before);
    twin.edit((choices) => renameTwin(choices, "plant-a"));
    twin.discard();
    expect(ready(twin).doc).toBe(before.doc);
    expect(ready(twin).edited).toBe(false);
    expect(hasUnsavedChoices(ready(twin))).toBe(false);
  });
});
