import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { reportWire } from "../../../test/runFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { liveWire } from "../../../test/liveFixture";
import { ReviewStore } from "./reviewStore";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(reportWire));
  const timers = fakeClock();
  const store = new ReviewStore(bridge.review, timers, timers);
  const known = sampleWorkspace.runs.find((run) => run.worldId === "w-2");
  if (known === undefined) throw new Error("missing fixture");
  store.open("w-2", { ...known, connected: "running" });
  const stream = (atNs: number) => {
    bridge.stream("w-2", [liveWire(atNs)], null);
    timers.advance(250);
  };
  return { store, bridge, stream, timers };
}

function originBoot(store: ReviewStore) {
  return store.snapshot().origins.get("zone-01")?.bootId;
}

function firstReading(store: ReviewStore) {
  return store.snapshot().live?.devices[0]?.readings[0]?.value;
}

describe("live review cursor", () => {
  it("plays hardware-only recordings across command, health and boot snapshots and rewinds the origin", () => {
    const { store, bridge, timers } = setup();
    const first = liveWire(10e9);
    const next = liveWire(11e9);
    const last = liveWire(12e9);
    bridge.stream("w-2", [
      first,
      {
        ...next,
        devices: next.devices.map((device) => ({
          ...device,
          boot_id: "62",
          reboots: 1,
          readings: device.readings.map((reading) => ({ ...reading, value: 32 })),
        })),
        commands: next.commands.map((command) => ({ ...command, receipt: "refused", reason: "local_control" })),
      },
      {
        ...last,
        host: "offline",
        devices: last.devices.map((device) => ({ ...device, boot_id: "62", health: "offline" })),
      },
      { type: "run.closed", seq: 13, stream: "observer", at_ns: 12e9, incomplete: false },
    ]);
    expect(store.index.bounds()).toBeNull();
    store.seek(10e9);
    expect(originBoot(store)).toBe("61");
    store.setSpeed(1);
    store.play();
    timers.advance(1000);
    expect(store.snapshot()).toMatchObject({
      positionNs: 11e9,
      playing: true,
      live: { commands: [{ receipt: "refused", reason: "local_control" }] },
    });
    expect(originBoot(store)).toBe("62");
    expect(firstReading(store)).toBe(32);
    timers.advance(1000);
    expect(store.snapshot()).toMatchObject({
      positionNs: 12e9,
      playing: false,
      live: { host: "offline", devices: [{ health: "offline" }] },
    });
    store.seek(10e9);
    expect(originBoot(store)).toBe("61");
    expect(firstReading(store)).toBe(28);
    store.play();
    timers.advance(1000);
    expect(store.snapshot()).toMatchObject({ live: { commands: [{ receipt: "refused" }] } });
    store.close();
  });

  it("batches live publication and keeps earlier published history immutable", () => {
    const { store, bridge, timers } = setup();
    let publications = 0;
    const stop = store.subscribe(() => {
      publications += 1;
    });
    bridge.stream(
      "w-2",
      Array.from({ length: 100 }, (_, index) => liveWire(index * 1e9)),
      null,
    );
    expect(publications).toBe(0);
    timers.advance(250);
    expect(publications).toBe(1);
    const history = store.snapshot().liveHistory;
    bridge.stream("w-2", [liveWire(100e9)], null);
    timers.advance(250);
    expect(store.snapshot().liveHistory?.previous).toBe(history);
    expect(history?.entries).toHaveLength(100);
    stop();
    store.close();
  });
  it("follows new records, keeps a scrubbed cursor across batches and resumes the latest snapshot", () => {
    const { store, stream } = setup();
    stream(10e9);
    expect(store.snapshot()).toMatchObject({ following: true, positionNs: 10e9 });
    store.seek(5e9);
    stream(20e9);
    expect(store.snapshot()).toMatchObject({ following: false, positionNs: 5e9, newestNs: 20e9, live: null });
    store.backToLive();
    expect(store.snapshot()).toMatchObject({ following: true, positionNs: 20e9, live: { atNs: 20e9 } });
    store.close();
  });

  it("turns a stopped connected session into ordinary review and ignores Back to live", () => {
    const { store, bridge, stream } = setup();
    stream(10e9);
    bridge.stream("w-2", [{ type: "run.closed", seq: 11, stream: "observer", at_ns: 11e9, incomplete: false }]);
    expect(store.snapshot()).toMatchObject({ connection: "saved", following: false });
    store.seek(2e9);
    store.backToLive();
    expect(store.snapshot().positionNs).toBe(2e9);
    store.close();
  });

  it("keeps a journalled origin where the live snapshot at the cursor does not hold the device yet", () => {
    const { store, bridge, timers } = setup();
    const early = { ...liveWire(5e9), devices: [] };
    const origin = {
      type: "origin",
      stream: "observer",
      seq: 4,
      at_ns: 6e9,
      instance: "zone-01",
      device: "sensor",
      board: "uno-07",
      boot_id: "61",
      environment: "hardware",
    };
    bridge.stream("w-2", [early, origin, liveWire(10e9)], null);
    timers.advance(250);
    store.seek(5e9);
    expect(store.snapshot().origins.get("zone-01")).toMatchObject({ kind: "real", board: "uno-07", bootId: "61" });
    store.close();
  });
});
