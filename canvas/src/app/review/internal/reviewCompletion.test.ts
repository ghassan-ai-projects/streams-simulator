import { describe, expect, it, vi } from "vitest";
import { success } from "../../ports";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { journal, report } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

const path = "runs/w-1/run.json";
const recorded = { id: "recorded-plant", title: "The plant recorded with the run" };
const pending = {
  worldId: "w-1",
  ended: false,
  complete: false,
  incompleteReason: "",
  explore: false,
  seed: 7,
  plant: "mini",
  durationS: 0,
  recordedAt: "2026-10-16T00:00:00Z",
  faults: 0,
};

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.verify", success({ matches: true }));
  bridge.addFile(path, { domain_spec: recorded }, "recorded");
  const timers = fakeClock();
  const store = new ReviewStore(bridge.review, timers, timers);
  return { bridge, store };
}

async function expectRecordedPlant(store: ReviewStore) {
  await vi.waitFor(() => {
    expect(store.snapshot().runDomain).toEqual({ phase: "ready", value: recorded });
  });
}

describe("reviewing a run listed before completion", () => {
  it("loads the recorded plant when the watched run closes", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", success({ ...report, ended: false, complete: false }));
    store.open("w-1", pending);
    await vi.waitFor(() => {
      expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: false } });
    });
    expect(bridge.reads).toEqual([]);
    bridge.actions.set("run.report", success(report));
    bridge.stream("w-1", journal);
    await expectRecordedPlant(store);
    expect(bridge.reads).toEqual([path]);
    expect(bridge.calls.filter((call) => call.name === "run.verify")).toHaveLength(1);
    store.close();
  });

  it("waits for artifact publication after the journal has already closed", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", success({ ...report, ended: false, complete: false }));
    store.open("w-1", pending);
    await vi.waitFor(() => {
      expect(store.snapshot().report.phase).toBe("ready");
    });
    bridge.stream("w-1", journal);
    await vi.waitFor(() => {
      expect(bridge.calls.filter((call) => call.name === "run.report")).toHaveLength(2);
    });
    expect(bridge.calls.some((call) => call.name === "run.verify")).toBe(false);
    bridge.actions.set("run.report", success(report));
    store.refreshCompleted("w-1", { ...pending, ended: true, complete: true });
    await expectRecordedPlant(store);
    expect(store.snapshot().verified).toMatchObject({ phase: "ready", value: { matches: true } });
    store.close();
  });

  it("uses the completed report when the listing is stale, without waiting for another stream close", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", success(report));
    store.open("w-1", pending);
    await expectRecordedPlant(store);
    expect(store.snapshot().verified).toMatchObject({ phase: "ready", value: { matches: true } });
    expect(bridge.reads).toEqual([path]);
    expect(bridge.calls.filter((call) => call.name === "run.verify")).toHaveLength(1);
    store.close();
  });

  it("reads an ended incomplete recording's plant without asking for an unavailable verdict", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", success({ ...report, complete: false }));
    store.open("w-1", { ...pending, ended: true });
    await expectRecordedPlant(store);
    expect(store.snapshot().verified).toMatchObject({
      phase: "failed",
      failure: { kind: "run-incomplete", diagnostic: { displayCode: "run_incomplete" } },
    });
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    store.close();
  });
});
