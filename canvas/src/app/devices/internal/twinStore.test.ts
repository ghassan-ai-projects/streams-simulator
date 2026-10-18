import { describe, expect, it, vi } from "vitest";
import { chooseInstrument, renameTwin, startAt } from "../../../model/devices";
import { flushPromises } from "../../../test/deferred";
import { cannedFailure, cannedFile } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import { failure } from "../../../test/outcome";
import { choicesPath, coolerPath, sensorDocument, sensorPath, twinBridge } from "../../../test/twinFixture";
import type { TwinDraft } from "./twinDraft";
import { canCreate, isChoicesDirty, othersOf, pairsOf } from "./twinSelectors";
import { createTwinSession } from "./twinStore";

function ready(twin: ReturnType<typeof createTwinSession>): TwinDraft {
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function opened(fake: FakeBridge = twinBridge()) {
  const twin = createTwinSession(fake.devices);
  await twin.open(coolerPath);
  return { twin, fake };
}

function kindDocument() {
  return { kind: "cooler", metrics: [], commands: [] };
}

describe("opening the behaviour of a device", () => {
  it("reads the device, its choices file, the other devices and the instrument library, and suggests by unit", async () => {
    const { twin, fake } = await opened();
    const draft = ready(twin);
    expect(draft.choicesPath).toBe(choicesPath);
    expect(draft.members.map((member) => member.kind.kind)).toEqual(["cooler"]);
    expect(draft.touched).toBe(false);
    expect(draft.saved).toBe(draft.doc);
    expect(draft.doc.kinds[0]?.metrics.temperature?.instrument).toBe("streamsim/thermistor@1");
    expect(draft.instruments?.map((instrument) => instrument.title)).toEqual([
      "Capacitive humidity sensor",
      "Tachometer",
      "Thermistor",
    ]);
    expect(othersOf(draft).map((entry) => entry.path)).toEqual([sensorPath]);
    expect(fake.calls.filter((call) => call.name === "component.show")).toHaveLength(3);
    expect(fake.reads).toContain(choicesPath);
    expect(pairsOf(draft)).toHaveLength(1);
    expect(isChoicesDirty(draft)).toBe(false);
  });

  it("reopens with the choices saved before, with the other devices they name", async () => {
    const fake = twinBridge();
    const first = await opened(fake);
    await first.twin.include(sensorPath);
    first.twin.edit((choices) => chooseInstrument(choices, "cooler", "temperature", "none"));
    first.twin.edit((choices) => startAt(choices, "cooler", "temperature", 20));
    await first.twin.persist();
    const draft = ready((await opened(fake)).twin);
    expect(draft.members.map((member) => member.kind.kind)).toEqual(["cooler", "sensor"]);
    expect(draft.touched).toBe(true);
    expect(draft.doc.kinds[0]?.metrics.temperature).toEqual({ instrument: "none", initial: 20 });
    expect(isChoicesDirty(draft)).toBe(false);
    expect(othersOf(draft)).toEqual([]);
  });

  it("adds what the interface gained since the choices were saved, which is then waiting to be saved", async () => {
    const fake = twinBridge();
    const first = await opened(fake);
    first.twin.edit((choices) => renameTwin(choices, "kept-twin"));
    await first.twin.persist();
    fake.addFile(coolerPath, { ...kindDocument(), metrics: [{ name: "pressure", data_type: "number" }] }, "k9");
    const draft = ready((await opened(fake)).twin);
    expect(draft.doc.id).toBe("kept-twin");
    expect(Object.keys(draft.doc.kinds[0]?.metrics ?? {})).toEqual(["pressure"]);
    expect(isChoicesDirty(draft)).toBe(true);
  });

  it("opens without the library or the device list", async () => {
    const fake = twinBridge();
    fake.actions.delete("component.list");
    fake.workspaceOutcome = failure(cannedFailure("boom", 500));
    const draft = ready((await opened(fake)).twin);
    expect(draft.instruments).toBeNull();
    expect(draft.known).toEqual([]);
    expect(draft.doc.kinds[0]?.metrics.temperature?.instrument).toBe("none");
  });

  it("keeps an instrument the binary cannot describe, without a unit", async () => {
    const fake = twinBridge();
    fake.handlers.set("component.show", () => failure(cannedFailure("not_found", 404)));
    const draft = ready((await opened(fake)).twin);
    expect(draft.instruments).toHaveLength(3);
    expect(draft.doc.kinds[0]?.metrics.temperature?.instrument).toBe("none");
  });

  it.each([
    ["the device is missing", (fake: FakeBridge) => fake.files.delete(coolerPath), "missing"],
    [
      "the device is not a kind file",
      (fake: FakeBridge) => {
        fake.addFile(coolerPath, "[1]");
      },
      "unreadable",
    ],
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

  it("ignores an answer that arrives after it was closed", async () => {
    const fake = twinBridge();
    const gates = fake.holdRead(coolerPath, 1);
    const twin = createTwinSession(fake.devices);
    const opening = twin.open(coolerPath);
    expect(twin.snapshot()).toEqual({ phase: "loading", path: coolerPath });
    await flushPromises();
    twin.close();
    gates[0]?.resolve(cannedFile(coolerPath, kindDocument(), "k1"));
    await opening;
    expect(twin.snapshot()).toEqual({ phase: "idle" });
    await twin.reload();
    expect(twin.snapshot()).toEqual({ phase: "idle" });
  });
});

describe("the other devices of a twin", () => {
  it("adds a device with its default choices and leaves it out again, never the first one", async () => {
    const { twin } = await opened();
    await twin.include(sensorPath);
    await twin.include(sensorPath);
    expect(ready(twin).doc.kinds.map((kind) => kind.kind)).toEqual(["cooler", "sensor"]);
    expect(ready(twin).touched).toBe(true);
    expect(othersOf(ready(twin))).toEqual([]);
    twin.leave("cooler");
    expect(ready(twin).members).toHaveLength(2);
    twin.leave("sensor");
    expect(ready(twin).doc.kinds.map((kind) => kind.kind)).toEqual(["cooler"]);
    expect(othersOf(ready(twin)).map((entry) => entry.path)).toEqual([sensorPath]);
  });

  it("does not add a device it cannot read, or one that was added while it waited", async () => {
    const { twin, fake } = await opened();
    fake.files.delete(sensorPath);
    await twin.include(sensorPath);
    expect(ready(twin).members).toHaveLength(1);
    fake.addFile(sensorPath, sensorDocument, "k2");
    const gates = fake.holdRead(sensorPath, 1);
    const including = twin.include(sensorPath);
    await flushPromises();
    twin.close();
    gates[0]?.resolve(cannedFile(sensorPath, sensorDocument, "k2"));
    await including;
    expect(twin.snapshot()).toEqual({ phase: "idle" });
  });
});

describe("saving the choices", () => {
  it("writes only what changed, one write at a time, each over the version the last one made", async () => {
    const { twin, fake } = await opened();
    await twin.persist();
    expect(fake.writes).toHaveLength(0);
    twin.edit((choices) => renameTwin(choices, "plant-a"));
    const first = twin.persist();
    twin.edit((choices) => renameTwin(choices, "plant-b"));
    const second = twin.persist();
    await Promise.all([first, second]);
    expect(fake.writes.map((write) => write.etag)).toEqual(["", "e1"]);
    expect(JSON.parse(fake.writes[1]?.content ?? "{}")).toMatchObject({ id: "plant-b" });
    expect(ready(twin).save).toEqual({ kind: "idle" });
    expect(isChoicesDirty(ready(twin))).toBe(false);
  });

  it("says when the file changed on disk and does not write over it", async () => {
    const { twin, fake } = await opened();
    fake.writeOutcomes.set(choicesPath, failure(cannedFailure("conflict", 409, "changed")));
    twin.edit((choices) => renameTwin(choices, "plant-a"));
    await twin.persist();
    expect(ready(twin).save).toMatchObject({ kind: "write-failed", failure: { kind: "conflict" } });
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

  it("keeps the state when an edit changes nothing", async () => {
    const { twin } = await opened();
    const before = ready(twin);
    twin.edit((choices) => choices);
    expect(ready(twin)).toBe(before);
  });
});
