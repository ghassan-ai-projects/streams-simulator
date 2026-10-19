import { describe, expect, it } from "vitest";
import { busyState, emptyState, liveOf, sensorDevice } from "../../../test/interruptFixture";
import {
  activeScopes,
  allDevices,
  chosenScope,
  interruptOf,
  newestInterrupts,
  scopeChoices,
  scopeStatus,
} from "./interruptSelect";

describe("which interrupt state is current", () => {
  it("prefers the answer of the node while no newer snapshot arrived, and the snapshot after", () => {
    const newest = liveOf({ key: "a", interrupts: emptyState });
    expect(newestInterrupts(newest, { state: busyState, seenKey: "a" })).toBe(busyState);
    expect(newestInterrupts(newest, { state: busyState, seenKey: "older" })).toBe(emptyState);
    expect(newestInterrupts(newest, null)).toBe(emptyState);
    expect(newestInterrupts(null, null)).toBeNull();
    expect(newestInterrupts(null, { state: busyState, seenKey: "" })).toBe(busyState);
    expect(newestInterrupts(liveOf({ interrupts: null }), null)).toBeNull();
  });

  it("finds what covers a device by its instance, through the device the snapshot names", () => {
    const live = liveOf();
    expect(interruptOf(live, "fan-01")).toEqual({ device: "fan", board: "mega-01", mode: "hold", latched: true });
    expect(interruptOf(live, "zone-01")).toMatchObject({ device: "sensor", mode: "hold", latched: false });
    expect(interruptOf(live, "pump-9")).toBeNull();
    expect(interruptOf(liveOf({ interrupts: null }), "fan-01")).toBeNull();
    expect(interruptOf(liveOf({ interrupts: emptyState }), "fan-01")).toBeNull();
    expect(interruptOf(null, "fan-01")).toBeNull();
  });
});

describe("scopes", () => {
  it("offers every device and each board the snapshot names, all first and sorted", () => {
    expect(scopeChoices(liveOf()).map((choice) => choice.label)).toEqual([
      "All devices",
      "Board mega-01",
      "Board uno-07",
      "Device fan",
      "Device sensor",
    ]);
    expect(scopeChoices(null).map((choice) => choice.key)).toEqual(["all:"]);
    const unnamed = liveOf({ devices: [{ ...sensorDevice, board: "" }] });
    expect(scopeChoices(unnamed).map((choice) => choice.key)).toEqual(["all:", "device:sensor"]);
  });

  it("finds the chosen scope and falls back to all devices", () => {
    const choices = scopeChoices(liveOf());
    expect(chosenScope(choices, "board:uno-07")).toEqual({ kind: "board", id: "uno-07" });
    expect(chosenScope(choices, "board:gone")).toBe(allDevices);
  });

  it("reads the hold, latch and approval of exactly one scope", () => {
    expect(scopeStatus(busyState, { kind: "device", id: "fan" })).toEqual({
      held: true,
      latched: true,
      approving: false,
    });
    expect(scopeStatus(busyState, { kind: "board", id: "uno-07" })).toEqual({
      held: true,
      latched: false,
      approving: false,
    });
    expect(scopeStatus(busyState, allDevices)).toEqual({ held: false, latched: false, approving: true });
    expect(scopeStatus(null, allDevices)).toEqual({ held: false, latched: false, approving: false });
  });

  it("lists the active holds, latches and approval modes in words", () => {
    expect(activeScopes(busyState).map((entry) => [entry.key, entry.text, entry.latched])).toEqual([
      ["hold:board:uno-07", "Board uno-07: Held", false],
      ["hold:device:fan", "Device fan: Held · latched by emergency stop", true],
      ["approve:all:", "All devices: Approving each command", false],
    ]);
    expect(activeScopes(null)).toEqual([]);
    expect(activeScopes(emptyState)).toEqual([]);
  });
});
