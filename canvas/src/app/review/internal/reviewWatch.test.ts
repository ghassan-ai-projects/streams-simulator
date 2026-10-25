import { describe, expect, it } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { observation } from "../../../test/runFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { ReviewStore } from "./reviewStore";

const startNs = 1_767_225_600e9;
const at = (second: number) => startNs + second * 1e9;

function setup() {
  const bridge = new FakeBridge();
  const timers = fakeClock();
  const store = new ReviewStore(bridge.review, timers, timers);
  const known = sampleWorkspace.runs.find((run) => run.worldId === "w-2");
  if (known === undefined) throw new Error("missing fixture");
  store.watch("w-2", { ...known, running: true });
  const send = (...entries: object[]) => {
    bridge.stream("w-2", entries, null);
    timers.advance(250);
  };
  return { bridge, store, timers, send };
}

const opened = {
  stream: "observer",
  type: "run.opened",
  seq: 0,
  at_ns: startNs,
  mode: "explore",
  adapter: { id: "a" },
};
const clock = (second: number) => ({
  stream: "observer",
  type: "clock.advanced",
  seq: 9,
  at_ns: at(second),
  to_ns: at(second),
});
const reading = (second: number, value: number, extra: object = {}) => {
  const wire = observation(1, at(second), "CT-1", "tower.basin_temp", value, "degC");
  return { ...wire, line: JSON.stringify({ ...JSON.parse(wire.line), ...extra }) };
};

describe("watching a run that is still going", () => {
  it("follows the newest reading and reads nothing but the observer journal", () => {
    const { bridge, store, send } = setup();
    send(opened, reading(10, 20.5), clock(10));
    send(reading(20, 21));
    expect(store.snapshot()).toMatchObject({
      worldId: "w-2",
      watching: true,
      following: true,
      running: true,
      connection: "none",
      positionNs: at(20),
    });
    expect(store.index.readings("CT-1", at(20))[0]).toMatchObject({ value: 21, atNs: at(20) });
    expect(bridge.calls).toEqual([]);
    expect(bridge.reads).toEqual([]);
    expect([...bridge.followers.keys()]).toEqual(["w-2"]);
    store.close();
  });

  it("keeps the run's start and the clock the world reached, even when no reading arrived", () => {
    const { store, send } = setup();
    send(opened, reading(10, 20));
    send(clock(70));
    expect(store.index.openedNs).toBe(startNs);
    expect(store.index.clockNs).toBe(at(70));
    store.close();
  });

  it("keeps each reading's quality and when it was measured", () => {
    const { store, send } = setup();
    send(reading(30, 20, { quality: "stale", event_time: "2026-01-01T00:00:10Z" }));
    expect(store.index.readings("CT-1", at(30))[0]).toMatchObject({ quality: "stale", lagNs: 20e9 });
    store.close();
  });

  it("has no report to wait for, and so no truth to read", () => {
    const { store, send } = setup();
    send(opened, reading(10, 20));
    expect(store.snapshot().report).toEqual({ phase: "loading" });
    store.close();
  });

  it("becomes an ordinary opened run when the journal closes, and stays on the newest reading", () => {
    const { bridge, store, send } = setup();
    send(opened, reading(10, 20), reading(20, 21));
    bridge.stream("w-2", [{ stream: "observer", type: "run.closed", seq: 99, at_ns: at(30), incomplete: false }]);
    expect(store.snapshot()).toMatchObject({ watching: false, following: false, positionNs: at(20) });
    send(reading(40, 99));
    expect(store.snapshot().positionNs).toBe(at(20));
    store.close();
  });

  it("stops watching when the stream ends without a close record, and does not jump to the start", () => {
    const { bridge, store, send } = setup();
    send(opened, reading(10, 20), reading(300, 21));
    bridge.stream("w-2", [], "closed");
    expect(store.snapshot()).toMatchObject({ watching: false, playback: { phase: "ready" } });
    expect(store.snapshot().positionNs).toBe(at(300));
    store.close();
  });

  it("reopens a stream that failed while the run was going", () => {
    const { bridge, store, send, timers } = setup();
    send(reading(10, 20));
    bridge.stream("w-2", [], "failed");
    expect(bridge.openFollows("w-2")).toBe(0);
    timers.advance(1000);
    expect(bridge.openFollows("w-2")).toBe(1);
    expect(store.snapshot()).toMatchObject({ watching: true, worldId: "w-2" });
    store.close();
  });

  it("is shared by two holders and released by the last one", () => {
    const { bridge, store } = setup();
    store.watch("w-2");
    store.close();
    expect(bridge.openFollows("w-2")).toBe(1);
    store.close();
    expect(bridge.openFollows("w-2")).toBe(0);
  });

  it("holds a bounded number of readings however long the run goes", () => {
    const { store, send } = setup();
    const entries = Array.from({ length: 6000 }, (_, second) => reading(second, second));
    send(...entries);
    const held = store.index.samples("CT-1", "tower.basin_temp", 0, 1e30)?.times.length ?? 0;
    expect(store.index.size).toBe(6000);
    expect(held).toBeLessThanOrEqual(4096);
    expect(held).toBeGreaterThanOrEqual(2048);
    expect(store.index.readings("CT-1", at(6000))[0]?.value).toBe(5999);
    store.close();
  });

  it("leaves an ordinary review of the same run unbounded", () => {
    const bridge = new FakeBridge();
    const timers = fakeClock();
    const store = new ReviewStore(bridge.review, timers, timers);
    store.open("w-1");
    bridge.stream(
      "w-1",
      Array.from({ length: 6000 }, (_, second) => reading(second, second)),
      null,
    );
    expect(store.index.samples("CT-1", "tower.basin_temp", 0, 1e30)?.times.length).toBe(6000);
    store.close();
  });
});
