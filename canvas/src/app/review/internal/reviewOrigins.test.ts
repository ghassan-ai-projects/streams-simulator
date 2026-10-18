import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { reportWire } from "../../../test/runFixture";
import { ReviewStore } from "./reviewStore";

function origin(instance: string, environment: string, bootId: string) {
  return {
    stream: "observer",
    seq: 1,
    type: "origin",
    at_ns: 0,
    instance,
    device: instance,
    board: "uno-07",
    environment,
    boot_id: bootId,
  };
}

describe("origins in a review", () => {
  it("keeps where each instance of a connected session came from, the latest boot winning", () => {
    const bridge = new FakeBridge();
    bridge.actions.set("run.report", success(reportWire));
    const timers = fakeClock();
    const store = new ReviewStore(bridge.review, timers, timers);
    store.open("c-1");
    bridge.stream(
      "c-1",
      [origin("zone-01", "hardware", "61"), origin("fan-01", "simulated", "b"), origin("zone-01", "hardware", "62")],
      null,
    );
    const origins = store.snapshot().origins;
    expect(origins.get("zone-01")).toEqual({
      instance: "zone-01",
      kind: "real",
      board: "uno-07",
      device: "zone-01",
      bootId: "62",
    });
    expect(origins.get("fan-01")?.kind).toBe("simulated");
    store.open("c-2");
    expect(store.snapshot().origins.size).toBe(0);
  });
});
