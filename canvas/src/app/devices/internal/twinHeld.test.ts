import { describe, expect, it, vi } from "vitest";
import { changeMetric } from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import { cannedFailure } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import {
  cooler,
  coolerDocument,
  coolerPath,
  kindOf,
  sensorDocument,
  sensorPath,
  twinBridge,
} from "../../../test/twinFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { createDeviceSession } from "./deviceStore";
import type { TwinDraft } from "./twinDraft";
import { createBlock } from "./twinSelectors";
import { createTwinSession } from "./twinStore";

const pumpPath = "kinds/pump_monitor.kind.json";
const unfinished = { kind: "pump_monitor", metrics: [{ name: "pressure", data_type: "number" }], commands: [] };
const finished = {
  kind: "pump_monitor",
  metrics: [{ name: "pressure", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, resolution: 1 }],
  commands: [],
};

function ready(twin: ReturnType<typeof createTwinSession>): TwinDraft {
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

function withPump(): FakeBridge {
  const fake = twinBridge();
  fake.addFile(pumpPath, unfinished, "k3");
  const listing = fake.workspaceOutcome;
  if (!listing.ok) throw new Error("no listing");
  const kinds = [...listing.value.kinds, { path: pumpPath, kind: "pump_monitor", title: "" }];
  fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
  return fake;
}

async function opened(fake: FakeBridge, lead: () => KindFile | null = () => cooler.kind) {
  const twin = createTwinSession(fake.devices, lead);
  await twin.open(coolerPath);
  return { twin, fake };
}

function suggestions(fake: FakeBridge) {
  return fake.calls.filter((call) => call.name === "kind.suggest").map((call) => call.args);
}

describe("a device that is not ready for a twin cannot hold another one back", () => {
  it("asks for the open device only, so an unfinished device elsewhere in the workspace changes nothing", async () => {
    const { twin, fake } = await opened(withPump());
    expect(ready(twin).held).toEqual([]);
    expect(ready(twin).doc.id).toBe("cooler-sensor-twin");
    expect(suggestions(fake)).toEqual([{ kinds: ["cooler"] }]);
  });

  it("names the unfinished device that is added, keeps it out, and adds it once it is ready", async () => {
    const { twin, fake } = await opened(withPump());
    await twin.include(pumpPath);
    expect(ready(twin).members.map((member) => member.path)).toEqual([coolerPath]);
    expect(ready(twin).held).toEqual([
      {
        path: pumpPath,
        name: "pump_monitor",
        words: "pump_monitor still needs a unit and a resolution",
        issues: [
          expect.objectContaining({
            owner: "reading",
            ownerName: "pressure",
            sentence: "Reading 'pressure' needs a unit.",
          }),
          expect.objectContaining({ sentence: "Reading 'pressure' needs a resolution." }),
        ] as unknown[],
        adding: true,
      },
    ]);
    expect(suggestions(fake).at(-1)).toEqual({ kinds: ["cooler", "pump_monitor"] });
    expect(createBlock(ready(twin), { phase: "idle" })).toBeNull();
    fake.addFile(pumpPath, finished, "k4");
    await twin.retry();
    expect(ready(twin).members.map((member) => member.path)).toEqual([coolerPath, pumpPath]);
    expect(ready(twin).held).toEqual([]);
    expect(ready(twin).doc.kinds).toHaveLength(2);
  });

  it("keeps a device it cannot read waiting, and drops it when the person leaves it out", async () => {
    const fake = withPump();
    fake.files.delete(pumpPath);
    const { twin } = await opened(fake);
    await twin.include(pumpPath);
    expect(ready(twin).held).toEqual([
      { path: pumpPath, name: "pump_monitor", words: "pump_monitor could not be read", issues: [], adding: true },
    ]);
    twin.leave(pumpPath);
    expect(ready(twin).held).toEqual([]);
  });

  it("names a device of a saved twin that stopped being ready, blocks Create, and recovers when it is fixed", async () => {
    const fake = twinBridge();
    const first = await opened(fake);
    await first.twin.include(sensorPath);
    await first.twin.persist();
    fake.addFile(sensorPath, { ...sensorDocument, metrics: [{ name: "humidity", data_type: "number" }] }, "k9");
    const { twin } = await opened(fake);
    expect(ready(twin).members).toHaveLength(2);
    expect(ready(twin).held).toEqual([
      {
        path: sensorPath,
        name: "sensor",
        words: "sensor still needs a unit and a resolution",
        issues: [
          expect.objectContaining({ ownerName: "humidity" }),
          expect.objectContaining({ ownerName: "humidity" }),
        ] as unknown[],
        adding: false,
      },
    ]);
    expect(createBlock(ready(twin), { phase: "idle" })).toBe("sensor still needs a unit and a resolution.");
    fake.addFile(sensorPath, sensorDocument, "k10");
    await twin.retry();
    expect(ready(twin).held).toEqual([]);
    expect(createBlock(ready(twin), { phase: "idle" })).toBeNull();
    expect(ready(twin).members).toHaveLength(2);
  });

  it("does nothing when there is nothing to retry", async () => {
    const { twin, fake } = await opened(twinBridge());
    const before = ready(twin);
    await twin.retry();
    expect(ready(twin)).toBe(before);
    expect(suggestions(fake)).toHaveLength(1);
  });
});

describe("opening the behaviour again after it could not be opened", () => {
  it("says the lead device is the one that lacks something, when it does", async () => {
    const fake = twinBridge();
    const lacking = { ...coolerDocument, metrics: [{ name: "pressure", data_type: "number", resolution: 1 }] };
    fake.addFile(coolerPath, lacking, "k1");
    const { twin } = await opened(fake, () => kindOf(lacking));
    expect(twin.snapshot()).toMatchObject({
      phase: "failed",
      held: [{ path: coolerPath, name: "cooler", words: "cooler still needs a unit", adding: false }],
    });
  });

  it("tries again from the failed state, also when another try is not needed", async () => {
    const fake = twinBridge();
    const good = fake.handlers.get("kind.suggest");
    fake.handlers.set("kind.suggest", () => failure(cannedFailure("boom", 500, "the bridge broke")));
    const { twin } = await opened(fake);
    expect(twin.snapshot()).toMatchObject({ phase: "failed", held: [], failure: { kind: "unavailable" } });
    if (good !== undefined) fake.handlers.set("kind.suggest", good);
    await twin.retry();
    expect(twin.snapshot().phase).toBe("ready");
  });
});

describe("the instruments a reading can use depend on its unit and its range", () => {
  it("asks again when the range of a reading changes, and keeps what the binary says of each instrument", async () => {
    let lead = cooler.kind;
    const { twin, fake } = await opened(twinBridge(), () => lead);
    lead = changeMetric(cooler.kind, "m1", { range: { min: 0, max: 200 } });
    await twin.open(coolerPath);
    const asked = fake.calls.filter((call) => call.name === "component.list").map((call) => call.args);
    expect(asked).toContainEqual({ unit: "Cel", min: 0, max: 200 });
    expect(ready(twin).instruments["Cel|0|200"]?.[0]?.fit).toEqual({
      ok: false,
      code: "range_narrower",
      member: "eu_range",
    });
    expect(ready(twin).instruments["Cel|0|100"]?.[0]?.fit).toEqual({ ok: true, code: "", member: "" });
  });
});

describe("the open device follows the workspace", () => {
  it("watches the workspace while it is open, retries a held twin when it changes, and stops when it is closed", async () => {
    const fake = withPump();
    const session = createDeviceSession(fake.devices, fake.workspace.watch);
    await session.open(coolerPath);
    expect(fake.watchers.size).toBe(1);
    await session.twin.include(pumpPath);
    fake.addFile(pumpPath, finished, "k4");
    fake.changeOnDisk();
    await vi.waitFor(() => {
      const state = session.twin.snapshot();
      expect(state.phase === "ready" && state.members.length).toBe(2);
    });
    session.close();
    expect(fake.watchers.size).toBe(0);
  });

  it("asks the binary about the twin again once the interface is saved, so its needs follow", async () => {
    const fake = twinBridge();
    const bare = { name: "temperature", data_type: "number", unit: "Cel", resolution: 0.1 };
    fake.addFile(coolerPath, { ...coolerDocument, metrics: [bare, ...coolerDocument.metrics.slice(1)] }, "k1");
    const session = createDeviceSession(fake.devices);
    await session.open(coolerPath);
    await vi.waitFor(() => {
      expect(session.twin.snapshot().phase).toBe("ready");
    });
    const before = suggestions(fake).length;
    session.edit((kind) => changeMetric(kind, "m1", { range: { min: 0, max: 100 } }));
    await session.save();
    await vi.waitFor(() => {
      const state = session.twin.snapshot();
      expect(state.phase === "ready" && state.instruments["Cel|0|100"]).toBeTruthy();
    });
    expect(suggestions(fake).length).toBeGreaterThan(before);
  });
});
