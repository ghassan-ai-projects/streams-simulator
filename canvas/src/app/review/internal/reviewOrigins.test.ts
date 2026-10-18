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
    expect(store.snapshot().connection).toBe("saved");
    store.open("c-2");
    expect(store.snapshot().origins.size).toBe(0);
    expect(store.snapshot().connection).toBe("none");
  });

  it("keeps the gateway binding and takes a connected session from the listing, running or saved", () => {
    const bridge = new FakeBridge();
    bridge.actions.set("run.report", success(reportWire));
    const timers = fakeClock();
    const store = new ReviewStore(bridge.review, timers, timers);
    const listed = {
      worldId: "c-3",
      ended: false,
      complete: false,
      incompleteReason: "",
      explore: false,
      seed: 1,
      plant: "p",
      durationS: 0,
      recordedAt: "",
      faults: 0,
      connected: "running" as const,
    };
    store.open("c-3", listed);
    expect(store.snapshot().connection).toBe("running");
    const binding = { stream: "observer", seq: 2, type: "binding", at_ns: 9, outcome: "refused", issues: [] };
    bridge.stream("c-3", [{ ...binding, refusal: "inventory_mismatch" }], null);
    expect(store.snapshot().binding).toEqual({
      atNs: 9,
      binding: { matched: false, refusal: "inventory_mismatch", reason: "", issues: [] },
    });
    expect(store.snapshot().connection).toBe("running");
    bridge.stream("c-3", [{ stream: "observer", seq: 3, type: "run.closed", at_ns: 10, incomplete: false }], null);
    expect(store.snapshot().connection).toBe("saved");
  });

  it("takes a stopped session from a later listing when the review opened on a stale one", () => {
    const bridge = new FakeBridge();
    bridge.actions.set("run.report", success(reportWire));
    const timers = fakeClock();
    const store = new ReviewStore(bridge.review, timers, timers);
    const listed = {
      worldId: "c-4",
      ended: false,
      complete: false,
      incompleteReason: "",
      explore: false,
      seed: 1,
      plant: "p",
      durationS: 0,
      recordedAt: "",
      faults: 0,
    };
    store.open("c-4", { ...listed, connected: "running" });
    store.refreshCompleted("c-4", { ...listed, ended: true, complete: true, connected: "saved" });
    expect(store.snapshot().connection).toBe("saved");
    store.refreshCompleted("c-9", { ...listed, connected: "running" });
    expect(store.snapshot().connection).toBe("saved");
  });
});
