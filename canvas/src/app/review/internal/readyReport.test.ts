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
  return { bridge, store: new ReviewStore(bridge.review, timers, timers) };
}

describe("ready report", () => {
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

  it("starts a run still going at the time its journal says it opened, not at its first reading", () => {
    const { store, bridge } = setup();
    store.watch("w-1");
    bridge.stream(
      "w-1",
      [
        { stream: "observer", type: "run.opened", seq: 0, at_ns: 100e9, mode: "explore", adapter: { id: "a" } },
        {
          stream: "observer",
          type: "observation",
          seq: 1,
          at_ns: 130e9,
          line: JSON.stringify({ entity_id: "A", channel: "a.b", value: 1, unit: "" }),
        },
      ],
      null,
    );
    expect(startNsOf(store.snapshot(), store)).toBe(100e9);
  });
});
