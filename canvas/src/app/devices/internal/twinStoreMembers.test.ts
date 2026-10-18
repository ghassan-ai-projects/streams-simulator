import { describe, expect, it } from "vitest";
import { addMetric, changeMetric, startAt } from "../../../model/devices";
import { flushPromises } from "../../../test/deferred";
import { cannedFile } from "../../../test/fakeBridge";
import { cooler, coolerPath, sensorDocument, sensorPath, twinBridge } from "../../../test/twinFixture";
import type { FakeBridge } from "../../../test/fakeBridge";
import type { KindFile } from "../../../model/devices";
import type { TwinDraft } from "./twinDraft";
import { hasUnsavedChoices, othersOf } from "./twinSelectors";
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

describe("coming back to the behaviour after the interface changed", () => {
  it("keeps what was chosen for a reading that was renamed, and asks the binary about a new one", async () => {
    let lead = cooler.kind;
    const { twin, fake } = await opened(twinBridge(), () => lead);
    twin.edit((choices) => startAt(choices, coolerPath, "m1", 30));
    const states: string[] = [];
    twin.subscribe(() => states.push(twin.snapshot().phase));
    lead = changeMetric(addMetric(changeMetric(cooler.kind, "m1", { name: "air" })).kind, "m4", {
      name: "humidity",
      unit: "%",
    });
    await twin.open(coolerPath);
    const draft = ready(twin);
    expect(draft.members[0]?.kind.metrics.map((metric) => metric.name)).toEqual(["air", "duty", "mode", "humidity"]);
    expect(draft.doc.kinds[0]?.metrics.m1?.initial).toBe(30);
    expect(Object.keys(draft.doc.kinds[0]?.metrics ?? {})).toEqual(["m1", "m2", "m3", "m4"]);
    expect(draft.edited).toBe(true);
    expect(states.every((phase) => phase === "ready")).toBe(true);
    expect(calls(fake, "kind.suggest")).toHaveLength(2);
    await twin.open(coolerPath);
    expect(calls(fake, "kind.suggest")).toHaveLength(2);
  });

  it("does not mark choices as edited by the user when only the interface changed", async () => {
    let lead = cooler.kind;
    const { twin } = await opened(twinBridge(), () => lead);
    lead = changeMetric(cooler.kind, "m1", { name: "air" });
    await twin.open(coolerPath);
    expect(ready(twin).edited).toBe(false);
    expect(hasUnsavedChoices(ready(twin))).toBe(false);
  });
});

describe("the other devices of a twin", () => {
  it("adds a device with its suggested choices and leaves it out again, never the first one", async () => {
    const { twin, fake } = await opened();
    await twin.include(sensorPath);
    await twin.include(sensorPath);
    expect(ready(twin).doc.kinds.map((kind) => kind.kind)).toEqual([coolerPath, sensorPath]);
    expect(ready(twin).doc.kinds[1]?.metrics.m1?.instrument).toBe("streamsim/humidity@1");
    expect(ready(twin).touched).toBe(true);
    expect(ready(twin).edited).toBe(true);
    expect(othersOf(ready(twin))).toEqual([]);
    expect(calls(fake, "component.list")).toHaveLength(2);
    twin.leave(coolerPath);
    expect(ready(twin).members).toHaveLength(2);
    twin.leave(sensorPath);
    expect(ready(twin).doc.kinds.map((kind) => kind.kind)).toEqual([coolerPath]);
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
