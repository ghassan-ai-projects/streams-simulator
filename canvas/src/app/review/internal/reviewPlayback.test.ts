import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { journalWire, reportWire } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(reportWire));
  bridge.actions.set("run.verify", success({ matches: true }));
  const timers = fakeClock();
  const skew = { ms: 0 };
  const store = new ReviewStore(bridge.review, { nowMs: () => timers.nowMs() + skew.ms }, timers);
  return { bridge, store, timers, skew };
}

function openRun(setupResult: ReturnType<typeof setup>, worldId = "w-1") {
  const { bridge, store, timers } = setupResult;
  store.open(worldId);
  bridge.stream(worldId, journalWire);
  timers.advance(250);
}

describe("review playback", () => {
  it("plays at its speed from the start, stops at the end and plays again from the start", () => {
    const run = setup();
    const { store, timers } = run;
    openRun(run);
    timers.advance(1000);
    expect(store.snapshot().positionNs).toBe(12e9);
    store.setSpeed(10);
    store.play();
    timers.advance(1000);
    expect(store.snapshot()).toMatchObject({ playing: true, positionNs: 22e9 });
    store.pause();
    timers.advance(1000);
    expect(store.snapshot()).toMatchObject({ playing: false, positionNs: 22e9 });
    store.play();
    timers.advance(5000);
    expect(store.snapshot()).toMatchObject({ playing: false, positionNs: 30e9 });
    store.play();
    expect(store.snapshot()).toMatchObject({ playing: true, positionNs: 12e9 });
    store.pause();
    store.pause();
    expect(store.snapshot().playing).toBe(false);
  });

  it("advances by the time the clock measured, not by the frame interval", () => {
    const run = setup();
    const { store, timers, skew } = run;
    openRun(run);
    store.setSpeed(1);
    store.play();
    skew.ms = 300;
    timers.advance(100);
    expect(store.snapshot().positionNs).toBe(12.4e9);
  });

  it("leaves no scheduled frame behind once playback pauses or reaches the end", () => {
    const run = setup();
    const { store, timers } = run;
    openRun(run);
    store.play();
    expect(timers.pending()).toBe(1);
    store.pause();
    expect(timers.pending()).toBe(0);
    store.play();
    store.play();
    expect(timers.pending()).toBe(1);
    timers.advance(60_000);
    expect(store.snapshot().playing).toBe(false);
    expect(timers.pending()).toBe(0);
  });

  it("stops the loop when the review closes, so nothing ticks afterwards", () => {
    const run = setup();
    const { store, timers } = run;
    openRun(run);
    store.play();
    timers.advance(100);
    store.close();
    const stopped = store.snapshot();
    expect(stopped.playing).toBe(false);
    expect(timers.pending()).toBe(0);
    timers.advance(10_000);
    expect(store.snapshot()).toBe(stopped);
  });

  it("does not leak a loop from one review into the next", () => {
    const run = setup();
    const { store, timers } = run;
    openRun(run, "w-1");
    store.play();
    store.open("w-2");
    expect(timers.pending()).toBe(0);
    openRun(run, "w-2");
    store.setSpeed(10);
    store.play();
    timers.advance(1000);
    expect(store.snapshot()).toMatchObject({ worldId: "w-2", playing: true, positionNs: 22e9 });
  });
});
