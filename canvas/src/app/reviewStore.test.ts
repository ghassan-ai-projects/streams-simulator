import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { FakeBridge, failure } from "../test/fakeBridge";
import { journal, report } from "../test/runFixture";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(report));
  bridge.actions.set("run.verify", success({ matches: true }));
  const flushes: (() => void)[] = [];
  const store = new ReviewStore(bridge, (flush) => flushes.push(flush));
  return { bridge, store, flushes };
}

describe("review", () => {
  it("draws the domain the run used from run.json, and says when it cannot", async () => {
    const { bridge, store } = setup();
    bridge.files.set(
      "runs/w-1/run.json",
      success({ path: "runs/w-1/run.json", content: JSON.stringify({ domain_spec: { id: "mini" } }), etag: "e" }),
    );
    store.open("w-1");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().runDomain).toEqual({ phase: "ready", value: { id: "mini" } });
    bridge.files.set("runs/w-2/run.json", success({ path: "runs/w-2/run.json", content: "{}", etag: "e" }));
    store.open("w-2");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().runDomain).toMatchObject({ phase: "failed", failure: { code: "no_domain_spec" } });
    store.open("w-3");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().runDomain).toMatchObject({ phase: "failed", failure: { code: "not_found" } });
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
    const { bridge, store, flushes } = setup();
    store.open("w-1");
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.calls.map((call) => [call.name, call.args])).toEqual([
      ["run.report", { run: "w-1" }],
      ["run.verify", { run: "w-1" }],
    ]);
    expect(store.snapshot()).toMatchObject({
      report: { phase: "ready", value: { domainId: "mini" } },
      verified: { phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } },
    });
    bridge.stream("w-1", journal);
    flushes.forEach((flush) => {
      flush();
    });
    expect(store.snapshot().playback).toEqual({ phase: "ready", records: 4, adapter: "native-jsonl" });
    expect(store.snapshot().positionNs).toBe(12e9);
    expect(store.index.readings("CH-1", 30e9)[0]?.value).toBe(95.1);
  });

  it("ignores Play before any observation has arrived and starts from where the run settles", () => {
    const { bridge, store, flushes } = setup();
    store.open("w-1");
    store.play();
    expect(store.snapshot().playing).toBe(false);
    bridge.stream("w-1", journal, null);
    flushes.forEach((flush) => {
      flush();
    });
    store.play();
    expect(store.snapshot()).toMatchObject({ playing: true, positionNs: 12e9 });
  });

  it("plays back at its speed from the start, stops at the end and plays again from the start", () => {
    const { bridge, store, flushes } = setup();
    store.open("w-1");
    bridge.stream("w-1", journal);
    flushes.forEach((flush) => {
      flush();
    });
    store.tick(1000);
    expect(store.snapshot().positionNs).toBe(12e9);
    store.setSpeed(10);
    store.play();
    store.tick(1000);
    expect(store.snapshot()).toMatchObject({ playing: true, positionNs: 22e9 });
    store.pause();
    store.tick(1000);
    expect(store.snapshot()).toMatchObject({ playing: false, positionNs: 22e9 });
    store.play();
    store.tick(5000);
    expect(store.snapshot()).toMatchObject({ playing: false, positionNs: 30e9 });
    store.play();
    expect(store.snapshot()).toMatchObject({ playing: true, positionNs: 12e9 });
    store.pause();
    store.pause();
    expect(store.snapshot().playing).toBe(false);
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

  it("drops late answers for a run that is no longer open and closes its stream", async () => {
    const { bridge, store } = setup();
    store.open("w-1");
    store.open("w-2");
    expect([...bridge.followers.keys()]).toEqual(["w-2"]);
    await Promise.resolve();
    await Promise.resolve();
    bridge.stream("w-1", journal);
    expect(store.snapshot()).toMatchObject({ worldId: "w-2", playback: { records: 0 } });
  });

  it("shows the binary's failure for report and verify, and a broken stream", async () => {
    const { bridge, store } = setup();
    bridge.actions.set("run.report", failed(failure("bridge_failed", 422, "no run.json")));
    bridge.actions.set("run.verify", failed(failure("bridge_failed", 422, "missing")));
    store.open("w-9");
    await Promise.resolve();
    await Promise.resolve();
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
      failed({ ...failure("replay_mismatch", 422, "no"), result: { matches: false, detail: "trace differs" } }),
    );
    store.open("w-1");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().verified).toEqual({
      phase: "ready",
      value: { matches: false, domainChanged: false, detail: "trace differs", reason: "", file: "" },
    });
    bridge.actions.set("run.verify", success({}));
    store.open("w-2");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().verified).toMatchObject({ phase: "failed", failure: { code: "verify_unreadable" } });
  });

  it("reads the report and verification again when a run it watched live closes, once", async () => {
    const { bridge, store, flushes } = setup();
    bridge.actions.set("run.report", success({ ...report, run: { ...report.run, ended: false } }));
    store.open("w-1");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: false } });
    bridge.actions.set("run.report", success(report));
    bridge.stream("w-1", journal);
    flushes.forEach((flush) => {
      flush();
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.calls.filter((call) => call.name === "run.report")).toHaveLength(2);
    expect(bridge.calls.filter((call) => call.name === "run.verify")).toHaveLength(2);
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: true } });
  });

  it("reviews a run the listing marks did-not-finish without asking for the doomed verdict and run.json", async () => {
    const { bridge, store } = setup();
    bridge.actions.set(
      "run.report",
      success({
        ...report,
        run: { ...report.run, ended: false, complete: false, incomplete_reason: "the run did not close." },
      }),
    );
    store.open("w-9", { ...entry, worldId: "w-9", ended: true, complete: false });
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    expect(bridge.reads).toEqual([]);
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: false } });
    expect(store.snapshot().verified).toMatchObject({
      phase: "failed",
      failure: { status: 422, code: "run_incomplete", message: "the run is incomplete: the run did not close." },
    });
    expect(store.snapshot().runDomain).toMatchObject({
      phase: "failed",
      failure: { status: 404, code: "not_found", message: "run.json is absent: the run is still open or did not end" },
    });
    expect(store.badgeFor("w-9")).toBe("incomplete");
  });

  it("answers a run the listing still shows as open from the listing, not the network", async () => {
    const { bridge, store } = setup();
    bridge.actions.set(
      "run.report",
      success({ ...report, run: { ...report.run, ended: false, complete: false, incomplete_reason: "" } }),
    );
    store.open("w-9", { ...entry, worldId: "w-9", ended: false, complete: false });
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    expect(bridge.reads).toEqual([]);
    expect(store.snapshot().verified).toMatchObject({
      phase: "failed",
      failure: { status: 422, code: "run_incomplete", message: "the run is incomplete: the run did not close." },
    });
    expect(store.snapshot().runDomain).toMatchObject({
      phase: "failed",
      failure: { status: 404, code: "not_found" },
    });
    expect(store.badgeFor("w-9")).toBe("incomplete");
  });

  it("finishes a run listed as open when its journal closes while the review watches", async () => {
    const { bridge, store, flushes } = setup();
    bridge.actions.set("run.report", success({ ...report, run: { ...report.run, ended: false, complete: false } }));
    store.open("w-1", { ...entry, worldId: "w-1", ended: false, complete: false });
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    bridge.actions.set("run.report", success(report));
    bridge.stream("w-1", journal);
    flushes.forEach((flush) => {
      flush();
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.calls.filter((call) => call.name === "run.report")).toHaveLength(2);
    expect(bridge.calls.filter((call) => call.name === "run.verify")).toHaveLength(1);
    expect(store.snapshot()).toMatchObject({
      report: { phase: "ready", value: { ended: true } },
      verified: { phase: "ready", value: { matches: true } },
    });
  });

  it("re-asks nothing doomed when an abandoned run's journal stream closes", async () => {
    const { bridge, store, flushes } = setup();
    store.open("w-9", { ...entry, worldId: "w-9", ended: true, complete: false });
    await Promise.resolve();
    await Promise.resolve();
    bridge.stream("w-9", journal);
    flushes.forEach((flush) => {
      flush();
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().playback).toMatchObject({ phase: "ready" });
    expect(bridge.calls.map((call) => call.name)).toEqual(["run.report"]);
    expect(bridge.reads).toEqual([]);
  });
});

const entry = {
  worldId: "w-9",
  ended: true,
  complete: true,
  incompleteReason: "the run did not close.",
  explore: false,
  seed: 5,
  plant: "factory-utilities",
  durationS: 0,
  recordedAt: "2026-10-14T10:00:00Z",
};
