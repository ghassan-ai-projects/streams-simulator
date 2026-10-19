import { describe, expect, it } from "vitest";
import { busyState, emptyState, liveOf, sensorDevice } from "../../../test/interruptFixture";
import {
  activeScopes,
  chosenChoice,
  interruptOf,
  newestInterrupts,
  scopeChoices,
  scopeStatus,
} from "./interruptSelect";
import { allDevices } from "./interruptScope";
import type { InterruptState } from "./interrupts";

const seen = (state: InterruptState, seenKey: string, older: readonly InterruptState[] = [], run = "run-1") => ({
  run,
  state,
  seenKey,
  older,
});

describe("which interrupt state is current", () => {
  it("prefers the answer of the node while no newer snapshot arrived, and the snapshot after", () => {
    const newest = liveOf({ key: "a", interrupts: emptyState });
    expect(newestInterrupts(newest, seen(busyState, "a"), "run-1")).toBe(busyState);
    expect(newestInterrupts(newest, seen(busyState, "older"), "run-1")).toBe(emptyState);
    expect(newestInterrupts(newest, null, "run-1")).toBe(emptyState);
    expect(newestInterrupts(null, null, "run-1")).toBeNull();
    expect(newestInterrupts(null, seen(busyState, ""), "run-1")).toBe(busyState);
    expect(newestInterrupts(liveOf({ interrupts: null }), null, "run-1")).toBeNull();
  });

  it("never shows what another run's control answered", () => {
    const newest = liveOf({ key: "a", interrupts: emptyState });
    expect(newestInterrupts(newest, seen(busyState, "a", [], "run-1"), "run-2")).toBe(emptyState);
    expect(newestInterrupts(null, seen(busyState, "", [], "run-1"), "run-2")).toBeNull();
    expect(newestInterrupts(null, seen(busyState, "", [], "run-1"), null)).toBeNull();
  });

  it("keeps the answer over a snapshot generated before the control but delivered after the answer", () => {
    const stale = liveOf({ key: "stale", interrupts: emptyState });
    const reply = seen(busyState, "a", [emptyState]);
    expect(newestInterrupts(stale, reply, "run-1")).toBe(busyState);
    const later = liveOf({ key: "later", interrupts: { ...busyState, pending: [] } });
    expect(newestInterrupts(later, reply, "run-1")).toBe(later.interrupts);
    const sameAsAnswer = liveOf({ key: "same", interrupts: busyState });
    expect(newestInterrupts(sameAsAnswer, reply, "run-1")).toBe(busyState);
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
    expect(scopeChoices(liveOf(), allDevices).map((choice) => choice.label)).toEqual([
      "All devices",
      "Board mega-01",
      "Board uno-07",
      "Device fan",
      "Device sensor",
    ]);
    expect(scopeChoices(null, allDevices).map((choice) => choice.key)).toEqual(["all:"]);
    const unnamed = liveOf({ devices: [{ ...sensorDevice, board: "" }] });
    expect(scopeChoices(unnamed, allDevices).map((choice) => choice.key)).toEqual(["all:", "device:sensor"]);
  });

  it("keeps a picked scope that left the session, labelled, instead of falling back to all devices", () => {
    const picked = { kind: "board", id: "gone" } as const;
    const choices = scopeChoices(liveOf(), picked);
    expect(choices.at(-1)).toEqual({
      key: "board:gone",
      scope: picked,
      label: "Board gone (not in this session)",
      present: false,
    });
    expect(chosenChoice(choices, picked)).toMatchObject({ scope: picked, present: false });
    expect(
      chosenChoice(scopeChoices(liveOf(), { kind: "board", id: "uno-07" }), { kind: "board", id: "uno-07" }),
    ).toMatchObject({ present: true });
    expect(chosenChoice([], allDevices)).toMatchObject({ key: "all:", present: true });
  });

  it("reads the hold, latch and approval of exactly one scope", () => {
    expect(scopeStatus(busyState, { kind: "device", id: "fan" })).toMatchObject({
      held: true,
      latched: true,
      approving: false,
    });
    expect(scopeStatus(busyState, { kind: "board", id: "uno-07" })).toMatchObject({
      held: true,
      latched: false,
      approving: false,
    });
    expect(scopeStatus(busyState, allDevices)).toMatchObject({ held: false, latched: false, approving: true });
    expect(scopeStatus(null, allDevices)).toEqual({ held: false, latched: false, approving: false, coveredBy: "" });
  });

  it("says which wider scope covers a device that has no hold of its own", () => {
    const sensor = { kind: "device", id: "sensor" } as const;
    expect(scopeStatus(busyState, sensor)).toMatchObject({ held: false, coveredBy: "Board uno-07: Held" });
    const led = { kind: "device", id: "led" } as const;
    expect(scopeStatus(busyState, led).coveredBy).toBe("All devices: Approving each command");
    const latchedAll: InterruptState = {
      ...busyState,
      holds: [{ scope: allDevices, latched: true }],
      approve: [],
    };
    expect(scopeStatus(latchedAll, sensor).coveredBy).toBe("All devices: Held · latched by emergency stop");
    expect(scopeStatus(busyState, { kind: "board", id: "mega-01" }).coveredBy).toBe(
      "All devices: Approving each command",
    );
    expect(scopeStatus(busyState, allDevices).coveredBy).toBe("");
  });

  it("lists the active holds, latches and approval modes in words", () => {
    expect(activeScopes(busyState).map((entry) => [entry.key, entry.label, entry.mode, entry.latched])).toEqual([
      ["hold:board:uno-07", "Board uno-07", "hold", false],
      ["hold:device:fan", "Device fan", "hold", true],
      ["approve:all:", "All devices", "approve", false],
    ]);
    expect(activeScopes(null)).toEqual([]);
    expect(activeScopes(emptyState)).toEqual([]);
  });
});
