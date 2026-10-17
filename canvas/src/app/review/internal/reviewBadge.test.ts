import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { fakeClock } from "../../../test/fakeClock";
import { journalWire, reportWire } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(reportWire));
  const verdict = deferred<Canned<unknown>>();
  bridge.handlers.set("run.verify", () => verdict.promise);
  const timers = fakeClock();
  return { bridge, verdict, timers, store: new ReviewStore(bridge.review, timers, timers) };
}

describe("review badge", () => {
  it("is recorded when the report and the verdict have both arrived, not before", async () => {
    const { verdict, store } = setup();
    store.open("w-1");
    await flushPromises();
    expect(store.snapshot().report.phase).toBe("ready");
    expect(store.badgeFor("w-1")).toBeNull();
    verdict.resolve(success({ matches: true }));
    await flushPromises();
    expect(store.badgeFor("w-1")).toBe("verified");
  });

  it("stays as recorded through playback, selection and reopening the run", async () => {
    const { bridge, verdict, timers, store } = setup();
    verdict.resolve(success({ matches: false }));
    store.open("w-1");
    await flushPromises();
    expect(store.badgeFor("w-1")).toBe("not-verified");
    bridge.stream("w-1", journalWire);
    timers.advance(250);
    store.select("CH-1");
    store.seek(20e9);
    expect(store.badgeFor("w-1")).toBe("not-verified");
    expect(store.badgeFor("w-2")).toBeNull();
  });
});
