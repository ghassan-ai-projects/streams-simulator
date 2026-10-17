import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { flushPromises } from "../../../test/deferred";
import { report, reportWire } from "../../../test/runFixture";
import { readyReport, startNsOf } from "./readyReport";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(reportWire));
  bridge.actions.set("run.verify", success({ matches: true }));
  const timers = fakeClock();
  return { store: new ReviewStore(bridge.review, timers, timers) };
}

describe("readyReport", () => {
  it("is null until the report has loaded, then the report", async () => {
    const { store } = setup();
    store.open("w-1");
    expect(readyReport(store.snapshot())).toBeNull();
    await flushPromises();
    expect(readyReport(store.snapshot())).toEqual(report);
  });

  it("starts the run at the report start once it is known", async () => {
    const { store } = setup();
    store.open("w-1");
    await flushPromises();
    expect(startNsOf(store.snapshot(), store)).toBe(report.startNs);
  });
});
