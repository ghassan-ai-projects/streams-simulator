import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { fakeClock } from "../../../test/fakeClock";
import { journalWire, reportWire, reportWith } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

const listed = {
  worldId: "w-1",
  ended: false,
  complete: false,
  running: false,
  incompleteReason: "",
  explore: false,
  seed: 5,
  plant: "mini",
  durationS: 0,
  recordedAt: "2026-10-14T10:00:00Z",
  faults: 0,
};

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.verify", success({ matches: true }));
  const gates: ReturnType<typeof deferred<Canned<unknown>>>[] = [];
  const answers: (Canned<unknown> | null)[] = [success(reportWith({ ended: false, complete: false }))];
  bridge.handlers.set("run.report", () => {
    const answer = answers.shift() ?? null;
    if (answer !== null) return answer;
    const gate = deferred<Canned<unknown>>();
    gates.push(gate);
    return gate.promise;
  });
  const timers = fakeClock();
  return { bridge, gates, store: new ReviewStore(bridge.review, timers, timers) };
}

describe("report ordering", () => {
  it("keeps the newest report when an earlier request answers last", async () => {
    const { bridge, gates, store } = setup();
    store.open("w-1", listed);
    await flushPromises();
    bridge.stream("w-1", journalWire);
    store.refreshCompleted("w-1", { ...listed, ended: true, complete: true });
    await flushPromises();
    expect(gates).toHaveLength(2);
    gates[1]?.resolve(success(reportWire));
    await flushPromises();
    gates[0]?.resolve(success(reportWith({ ended: false })));
    await flushPromises();
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: true } });
    expect(store.badgeFor("w-1")).not.toBe("incomplete");
  });

  it("applies a report that is still the newest when it answers", async () => {
    const { bridge, gates, store } = setup();
    store.open("w-1", listed);
    await flushPromises();
    bridge.stream("w-1", journalWire);
    await flushPromises();
    gates[0]?.resolve(success(reportWire));
    await flushPromises();
    expect(store.snapshot().report).toMatchObject({ phase: "ready", value: { ended: true } });
  });
});
