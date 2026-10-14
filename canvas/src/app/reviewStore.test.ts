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
    store.reveal(600e9, "CT-1");
    expect(store.snapshot()).toMatchObject({ positionNs: 600e9, selected: "CT-1", tab: "report" });
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
      verified: { phase: "ready", value: { matches: true, domainChanged: false, detail: "" } },
    });
    bridge.stream("w-1", journal);
    flushes.forEach((flush) => {
      flush();
    });
    expect(store.snapshot().playback).toEqual({ phase: "ready", records: 4, adapter: "native-jsonl" });
    expect(store.snapshot().positionNs).toBe(30e9);
    expect(store.index.readings("CH-1", store.snapshot().positionNs)[0]?.value).toBe(95.1);
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
      value: { matches: false, domainChanged: false, detail: "trace differs" },
    });
    bridge.actions.set("run.verify", success({}));
    store.open("w-2");
    await Promise.resolve();
    await Promise.resolve();
    expect(store.snapshot().verified).toMatchObject({ phase: "failed", failure: { code: "verify_unreadable" } });
  });
});
