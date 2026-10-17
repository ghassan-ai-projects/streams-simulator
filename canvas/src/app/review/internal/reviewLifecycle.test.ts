import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { deferred, settle } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { fakeClock } from "../../../test/fakeClock";
import { journal, report } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(report));
  bridge.actions.set("run.verify", success({ matches: true }));
  const timers = fakeClock();
  const store = new ReviewStore(bridge.review, timers, timers);
  return { bridge, store, timers };
}

describe("review lifecycle", () => {
  it("drops late answers for a run that is no longer open and closes its stream", async () => {
    const { bridge, store } = setup();
    store.open("w-1");
    store.open("w-2");
    expect([...bridge.followers.keys()]).toEqual(["w-2"]);
    await settle();
    bridge.stream("w-1", journal);
    expect(store.snapshot()).toMatchObject({ worldId: "w-2", playback: { records: 0 } });
  });

  it("ignores a flushed update whose session has already been superseded", async () => {
    const { bridge, store, timers } = setup();
    store.open("w-1");
    bridge.streamFollow(0, [{ kind: "opened", atNs: 0, mode: "", adapter: "wire" }]);
    store.open("w-2");
    timers.advance(250);
    await Promise.resolve();
    expect(store.snapshot().playback.adapter).toBe("");
  });

  it("ignores a retained callback of a superseded run after another run opened", async () => {
    const { bridge, store, timers } = setup();
    store.open("w-1");
    store.open("w-2");
    bridge.streamFollow(0, journal);
    timers.advance(250);
    await Promise.resolve();
    expect(store.snapshot()).toMatchObject({ worldId: "w-2", playback: { records: 0 } });
  });

  it("ignores a retained callback of an earlier visit to the same run", async () => {
    const { bridge, store, timers } = setup();
    store.open("w-1");
    store.open("w-2");
    store.open("w-1");
    bridge.streamFollow(0, journal);
    timers.advance(250);
    await Promise.resolve();
    expect(store.snapshot()).toMatchObject({ worldId: "w-1", playback: { records: 0 } });
  });

  it("does not publish anything for a run that is no longer open", async () => {
    const { bridge, store } = setup();
    store.open("w-1");
    await settle();
    store.close();
    bridge.actions.set("run.report", success({ ...report, ended: false }));
    store.refreshCompleted("w-1", { ...entry, ended: true, complete: false });
    await settle();
    expect(bridge.calls.filter((call) => call.name === "run.report")).toHaveLength(1);
  });

  it("does not publish a report that arrives after the same run was reopened", async () => {
    const { bridge, store } = setup();
    const gate = deferred<Canned<unknown>>();
    let holding = true;
    bridge.handlers.set("run.report", () => {
      if (!holding) return success(report);
      holding = false;
      return gate.promise;
    });
    store.open("w-1");
    store.open("w-2");
    store.open("w-1");
    gate.resolve(success({ ...report, ended: false }));
    await settle();
    expect(store.snapshot()).toMatchObject({ worldId: "w-1", report: { phase: "ready", value: { ended: true } } });
  });

  const entry = {
    worldId: "w-1",
    ended: true,
    complete: true,
    incompleteReason: "",
    explore: false,
    seed: 5,
    plant: "factory-utilities",
    durationS: 0,
    recordedAt: "2026-10-14T10:00:00Z",
  };
});
