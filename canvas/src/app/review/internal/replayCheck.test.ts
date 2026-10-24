import { describe, expect, it } from "vitest";
import { flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { reportWire } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";
import { fakeClock } from "../../../test/fakeClock";

const checked = { ...reportWire, replay: { status: "reproduces" } };

function setup() {
  const bridge = new FakeBridge();
  const asked: Record<string, unknown>[] = [];
  bridge.handlers.set("run.report", (args) => {
    asked.push({ ...args });
    return success(args.replay === true ? checked : reportWire);
  });
  bridge.actions.set("run.verify", success({ matches: true }));
  const timers = fakeClock();
  const store = new ReviewStore(bridge.review, timers, timers);
  return { bridge, asked, store };
}

describe("checking the replay of a run from its report", () => {
  it("starts idle with a report that was not replayed", async () => {
    const { store } = setup();
    store.open("w-1");
    await flushPromises();
    expect(store.snapshot().replayCheck).toEqual({ phase: "idle" });
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { replay: { status: "not_checked" } } });
  });

  it("asks for the report again with the replay and shows the new replay status", async () => {
    const { store, asked } = setup();
    store.open("w-1");
    await flushPromises();
    store.checkReplay();
    expect(store.snapshot().replayCheck).toEqual({ phase: "checking" });
    await flushPromises();
    expect(asked.at(-1)).toEqual({ run: "w-1", replay: true });
    expect(store.snapshot().replayCheck).toEqual({ phase: "idle" });
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { replay: { status: "reproduces" } } });
  });

  it("keeps the report and shows the binary's failure word for word", async () => {
    const { bridge, store } = setup();
    store.open("w-1");
    await flushPromises();
    bridge.handlers.set("run.report", () =>
      failure(cannedFailure("replay_failed", 500, "the recording is unreadable")),
    );
    store.checkReplay();
    await flushPromises();
    expect(store.snapshot().replayCheck).toEqual({ phase: "failed", message: "the recording is unreadable" });
    expect(store.snapshot().report.phase).toBe("ready");
  });

  it("does nothing when no run is open or a check is already running", async () => {
    const { store, asked } = setup();
    store.checkReplay();
    expect(asked).toEqual([]);
    store.open("w-1");
    await flushPromises();
    store.checkReplay();
    store.checkReplay();
    await flushPromises();
    expect(asked.filter((args) => args.replay === true)).toHaveLength(1);
  });

  it("forgets a check when another run is opened", async () => {
    const { store } = setup();
    store.open("w-1");
    await flushPromises();
    store.checkReplay();
    store.close();
    store.open("w-2");
    await flushPromises();
    expect(store.snapshot().replayCheck).toEqual({ phase: "idle" });
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { replay: { status: "not_checked" } } });
  });
});
