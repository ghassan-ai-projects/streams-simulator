import { describe, expect, it } from "vitest";
import type { InstanceOrigin } from "../../graph";
import { plantView } from "../../graph";
import { parsePlant } from "../../plant";
import type { LiveFacts } from "./live";
import { liveOf } from "./live";
import { PlaybackIndex } from "./playback";
import { reviewView } from "./reviewRows";
import type { SessionDevice, SessionLive, SessionReading } from "./sessionLive";

const bench = parsePlant({
  id: "bench",
  title: "Bench",
  site: { name: "Bench", areas: [{ name: "Bench" }] },
  kinds: [
    { name: "fan", channels: [{ name: "duty_permille", unit: "permille" }] },
    {
      name: "indicator",
      channels: [
        { name: "brightness_permille", unit: "permille" },
        { name: "pattern", unit: "code" },
      ],
    },
    { name: "sensor", channels: [{ name: "temperature", unit: "Cel" }] },
  ],
  instances: [
    { kind: "fan", names: ["fan-01"], area: "Bench" },
    { kind: "indicator", names: ["led-01"], area: "Bench" },
    { kind: "sensor", names: ["zone-01"], area: "Bench" },
  ],
  links: [],
});

const view = plantView(bench, new Map(), { open: null, view: reviewView });

function reading(name: string, unit: string, value: SessionReading["value"], state: SessionReading["state"] = "good") {
  return { name, unit, value, quality: "good", state } satisfies SessionReading;
}

function device(instance: string, readings: readonly SessionReading[]): SessionDevice {
  return {
    device: instance,
    board: "board-01",
    instance,
    environment: "hardware",
    bootId: "boot-72",
    health: "online",
    reboots: 0,
    readings,
  };
}

function sessionLive(devices: readonly SessionDevice[]): SessionLive {
  return {
    key: "k",
    commandsTruncated: false,
    marks: [],
    atNs: 1e9,
    host: "online",
    gateway: "online",
    devices,
    commands: [],
    interrupts: null,
  };
}

const origins = new Map<string, InstanceOrigin>(
  ["fan-01", "led-01", "zone-01"].map((instance) => [
    instance,
    { instance, kind: "real", board: "board-01", device: instance, bootId: "boot-72" },
  ]),
);

function facts(devices: readonly SessionDevice[]): LiveFacts {
  return {
    index: new PlaybackIndex(),
    plant: bench,
    view,
    faulted: new Map(),
    effectors: new Map(),
    origins,
    positionNs: 1e9,
    sessionLive: sessionLive(devices),
  };
}

const board = [
  device("fan-01", [reading("level_percent", "%", 45)]),
  device("led-01", [reading("state", "", "alert")]),
  device("zone-01", [reading("temperature", "Cel", 27.4)]),
];

function rowsOf(devices: readonly SessionDevice[], id: string) {
  return liveOf(facts(devices), id).readout.rows.map((row) => [row.label, row.value, row.unit, row.present] as const);
}

describe("a real board's node box", () => {
  it("shows the kind's metrics, not the domain's channels", () => {
    expect(rowsOf(board, "fan-01")).toEqual([["level percent", "45.0", "%", true]]);
    expect(rowsOf(board, "led-01")).toEqual([["state", "alert", "", true]]);
    expect(rowsOf(board, "zone-01")).toEqual([["temperature", "27.4", "°C", true]]);
  });

  it("shows a reading that carries no value as its state, and keeps a zero", () => {
    const quiet = [device("fan-01", [reading("level_percent", "%", 0)])];
    expect(rowsOf(quiet, "fan-01")[0]).toEqual(["level percent", "0.00", "%", true]);
    const lost = [device("led-01", [reading("state", "", "off", "disconnected")])];
    expect(rowsOf(lost, "led-01")).toEqual([["state", "disconnected", "", true]]);
  });

  it("keeps to the rows the layout reserved and counts the rest as more", () => {
    const many = [
      device("led-01", [reading("state", "", "off"), reading("a", "", 1), reading("b", "", 2), reading("c", "", 3)]),
    ];
    const { readout } = liveOf(facts(many), "led-01");
    expect(readout.rows.map((row) => row.label)).toEqual(["state"]);
    expect(readout.hidden).toBe(3);
  });

  it("falls back to the domain's channels until the board has a device to read", () => {
    const { readout } = liveOf(facts([]), "fan-01");
    expect(readout.rows.map((row) => [row.label, row.present])).toEqual([["duty permille", false]]);
  });

  it("tells the status from the device's readings only", () => {
    expect(liveOf(facts(board), "fan-01")).toMatchObject({ status: null });
    expect(liveOf(facts(board), "fan-01").shown.map((shown) => shown.channel)).toEqual(["level_percent"]);
  });
});
