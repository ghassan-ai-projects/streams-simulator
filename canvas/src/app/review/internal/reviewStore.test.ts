import { flushPromises } from "../../../test/deferred";
import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../ports";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { journalWire, reportWire, reportWith } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(reportWire));
  bridge.actions.set("run.verify", success({ matches: true }));
  const timers = fakeClock();
  const store = new ReviewStore(bridge.review, timers, timers);
  return { bridge, store, timers };
}

describe("review store", () => {
  it("draws the domain the run used from run.json, and says when it cannot", async () => {
    const { bridge, store } = setup();
    bridge.addFile("runs/w-1/run.json", { domain_spec: { id: "mini" } });
    store.open("w-1");
    await flushPromises();
    expect(store.snapshot().runDomain).toEqual({ phase: "ready", value: { id: "mini" } });
    bridge.addFile("runs/w-2/run.json", "{}");
    store.open("w-2");
    await flushPromises();
    expect(store.snapshot().runDomain).toMatchObject({ phase: "failed", failure: { kind: "unreadable" } });
    store.open("w-3");
    await flushPromises();
    expect(store.snapshot().runDomain).toMatchObject({
      phase: "failed",
      failure: { kind: "missing", diagnostic: { displayCode: "not_found" } },
    });
  });

  it("opens the inspector on a selection and keeps the chosen tab otherwise", () => {
    const { store } = setup();
    store.open("w-1");
    expect(store.snapshot().tab).toBe("report");
    store.select("CH-1");
    expect(store.snapshot().tab).toBe("inspect");
    store.showTab("report");
    store.select(null);
    expect(store.snapshot().tab).toBe("report");
  });

  it("reveals a mark by seeking and selecting without changing the tab, on either tab", () => {
    const { store } = setup();
    store.open("w-1");
    store.reveal(600e9, "CT-1");
    expect(store.snapshot()).toMatchObject({
      positionNs: 600e9,
      selected: "CT-1",
      tab: "report",
      picked: { atNs: 600e9, entity: "CT-1" },
    });
    store.showTab("inspect");
    store.reveal(900e9, "CH-1");
    expect(store.snapshot()).toMatchObject({ positionNs: 900e9, selected: "CH-1", tab: "inspect" });
  });

  it("asks the binary for the report and verification and streams the observer journal", async () => {
    const { bridge, store, timers } = setup();
    store.open("w-1");
    await flushPromises();
    expect(bridge.calls.map((call) => [call.name, call.args])).toEqual([
      ["run.report", { run: "w-1" }],
      ["run.verify", { run: "w-1" }],
    ]);
    expect(store.snapshot()).toMatchObject({
      report: { phase: "ready", value: { domainId: "mini" } },
      verified: { phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } },
    });
    bridge.stream("w-1", journalWire);
    timers.advance(250);
    expect(store.snapshot().playback).toEqual({ phase: "ready", records: 4 });
    expect(store.snapshot().positionNs).toBe(12e9);
    expect(store.index.readings("CH-1", 30e9)[0]?.value).toBe(95.1);
  });

  it("ignores Play before any observation has arrived and starts from where the run settles", () => {
    const { bridge, store, timers } = setup();
    store.open("w-1");
    store.play();
    expect(store.snapshot().playing).toBe(false);
    bridge.stream("w-1", journalWire, null);
    timers.advance(250);
    store.play();
    expect(store.snapshot()).toMatchObject({ playing: true, positionNs: 12e9 });
  });

  it("seeks, selects and ignores reopening the same run", () => {
    const { bridge, store } = setup();
    store.open("w-1");
    store.seek(15e9);
    store.select("CH-1");
    store.open("w-1");
    expect(store.snapshot()).toMatchObject({ positionNs: 15e9, selected: "CH-1" });
    expect(bridge.followers.size).toBe(1);
  });

  it("shows the binary's failure for report and verify, and a broken stream", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", failed(cannedFailure("bridge_failed", 422, "no run.json")));
    bridge.actions.set("run.verify", failed(cannedFailure("bridge_failed", 422, "missing")));
    store.open("w-9");
    await flushPromises();
    bridge.stream("w-9", [], "failed");
    expect(store.snapshot()).toMatchObject({
      report: { phase: "failed" },
      verified: { phase: "failed" },
      playback: { phase: "failed" },
    });
    store.close();
    expect(bridge.followers.size).toBe(0);
  });

  it("reads a verify that exits with a divergence as its verdict, and a verdict-less success as a failure", async () => {
    const { bridge, store } = setup();
    bridge.actions.set(
      "run.verify",
      failed({ ...cannedFailure("replay_mismatch", 422, "no"), result: { matches: false, detail: "trace differs" } }),
    );
    store.open("w-1");
    await flushPromises();
    expect(store.snapshot().verified).toEqual({
      phase: "ready",
      value: { matches: false, domainChanged: false, detail: "trace differs", reason: "", file: "" },
    });
    bridge.actions.set("run.verify", success({}));
    store.open("w-2");
    await flushPromises();
    expect(store.snapshot().verified).toMatchObject({ phase: "failed", failure: { kind: "unreadable" } });
  });

  it("reads the report and verification again when a run it watched live closes, once", async () => {
    const { bridge, store, timers } = setup();
    bridge.actions.set("run.report", success(reportWith({ ended: false })));
    store.open("w-1");
    await flushPromises();
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: false } });
    bridge.actions.set("run.report", success(reportWire));
    bridge.stream("w-1", journalWire);
    timers.advance(250);
    await flushPromises();
    expect(bridge.calls.filter((call) => call.name === "run.report")).toHaveLength(2);
    expect(bridge.calls.filter((call) => call.name === "run.verify")).toHaveLength(2);
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: true } });
  });

  it("reviews a run the listing marks did-not-finish without asking for the doomed verdict and run.json", async () => {
    const { bridge, store } = setup();
    bridge.actions.set(
      "run.report",
      success(reportWith({ ended: false, complete: false, incomplete_reason: "the run did not close." })),
    );
    store.open("w-9", { ...entry, worldId: "w-9", ended: true, complete: false });
    await flushPromises();
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    expect(bridge.reads).toEqual([]);
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: false } });
    expect(store.snapshot().verified).toMatchObject({
      phase: "failed",
      failure: {
        kind: "run-incomplete",
        diagnostic: { displayCode: "run_incomplete", message: "the run is incomplete: the run did not close." },
      },
    });
    expect(store.snapshot().runDomain).toMatchObject({
      phase: "failed",
      failure: {
        kind: "run-incomplete",
        diagnostic: {
          displayCode: "run_incomplete",
          message: "run.json is absent: the run is still open or did not end",
        },
      },
    });
    expect(store.badgeFor("w-9")).toBe("incomplete");
  });

  it("answers a run the listing still shows as open from the listing, not the network", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", success(reportWith({ ended: false, complete: false, incomplete_reason: "" })));
    store.open("w-9", { ...entry, worldId: "w-9", ended: false, complete: false });
    await flushPromises();
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    expect(bridge.reads).toEqual([]);
    expect(store.snapshot().verified).toMatchObject({
      phase: "failed",
      failure: {
        kind: "run-incomplete",
        diagnostic: { displayCode: "run_incomplete", message: "the run is incomplete: the run did not close." },
      },
    });
    expect(store.snapshot().runDomain).toMatchObject({
      phase: "failed",
      failure: { kind: "run-incomplete", diagnostic: { displayCode: "run_incomplete" } },
    });
    expect(store.badgeFor("w-9")).toBe("incomplete");
  });

  it("re-asks nothing doomed when an abandoned run's journal stream closes", async () => {
    const { bridge, store, timers } = setup();
    bridge.actions.set("run.report", success(reportWith({ ended: false, complete: false })));
    store.open("w-9", { ...entry, worldId: "w-9", ended: true, complete: false });
    await flushPromises();
    bridge.stream("w-9", journalWire);
    timers.advance(250);
    await flushPromises();
    expect(store.snapshot().playback).toMatchObject({ phase: "ready" });
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    expect(bridge.reads).toEqual([]);
  });
});

const entry = {
  worldId: "w-9",
  ended: true,
  complete: true,
  running: false,
  incompleteReason: "the run did not close.",
  explore: false,
  seed: 5,
  plant: "factory-utilities",
  durationS: 0,
  recordedAt: "2026-10-14T10:00:00Z",
  faults: 0,
};
