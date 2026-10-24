import { describe, expect, it } from "vitest";
import { blankReport, report } from "../../../test/runFixture";
import type { OperatorCall } from "./callLog";
import { eventLine, overlayOf, truthAllowed } from "./trendOverlay";

const whole = { startNs: 0, endNs: 3_600e9 };

const calls: readonly OperatorCall[] = [
  { atNs: 100e9, entity: "CH-1", effector: "set_setpoint", accepted: true, reason: "" },
  { atNs: 200e9, entity: "CT-1", effector: "set_fan_speed", accepted: false, reason: "interlock" },
];

describe("what is drawn on the shared time axis besides the readings", () => {
  it("always shows the operator's own effector calls, accepted and refused, in time order", () => {
    const blinded = blankReport({ mode: "evaluate", ended: false, truthUnavailable: "sealed until the run ends" });
    const events = overlayOf(calls, blinded, whole);
    expect(events.map((event) => [event.kind, event.entity, event.refused])).toEqual([
      ["effector", "CH-1", false],
      ["effector", "CT-1", true],
    ]);
    expect(events[0]?.label).toBe("CH-1 set setpoint · accepted");
    expect(events[1]?.label).toBe("CT-1 set fan speed · refused");
  });

  it("shows no fault, and nothing from the truth report, while an Evaluate run is open", () => {
    const open = { ...report, ended: false, truthUnavailable: "" };
    expect(truthAllowed(open)).toBe(false);
    const events = overlayOf(calls, open, whole);
    expect(events.every((event) => event.kind === "effector")).toBe(true);
    expect(events).toHaveLength(2);
    expect(JSON.stringify(events)).not.toMatch(/fan.failure/i);
  });

  it("shows the injected faults of an Explore run while it runs", () => {
    const explore = { ...report, mode: "explore", ended: false };
    expect(truthAllowed(explore)).toBe(true);
    const faults = overlayOf([], explore, whole).filter((event) => event.kind === "fault");
    expect(faults.map((event) => [event.entity, event.atNs])).toEqual([["CT-1", 600e9]]);
    expect(faults[0]?.label).toBe("CT-1 fan failure");
  });

  it("shows the faults of a finished Evaluate run once it is sealed, but not when the truth is unavailable", () => {
    expect(overlayOf([], report, whole).some((event) => event.kind === "fault")).toBe(true);
    expect(truthAllowed({ ...report, truthUnavailable: "files of another run" })).toBe(false);
    expect(truthAllowed(null)).toBe(false);
    expect(overlayOf([], null, whole)).toEqual([]);
  });

  it("adds the calls a sealed report knows and the journal did not, without doubling the ones it did", () => {
    const journalled: OperatorCall[] = [
      { atNs: 1_200e9, entity: "CH-1", effector: "start", accepted: true, reason: "" },
    ];
    const events = overlayOf(journalled, report, whole).filter((event) => event.kind === "effector");
    expect(events.map((event) => [event.atNs, event.refused])).toEqual([
      [1_200e9, false],
      [1_300e9, true],
      [1_400e9, false],
      [1_500e9, true],
    ]);
    expect(events.map((event) => event.label).join("|")).not.toMatch(/no effect|silent/);
  });

  it("keeps only the events inside the window", () => {
    const events = overlayOf(calls, null, { startNs: 150e9, endNs: 300e9 });
    expect(events.map((event) => event.entity)).toEqual(["CT-1"]);
  });

  it("words an event with its run time and kind", () => {
    const [event] = overlayOf([], { ...report, mode: "explore" }, whole);
    expect(event).toBeDefined();
    const fault = overlayOf([], { ...report, mode: "explore" }, whole).find((entry) => entry.kind === "fault");
    expect(fault === undefined ? "" : eventLine(fault, 0)).toBe("00:10:00 · Fault starts · CT-1 fan failure");
  });

  it("adds the faults of an Explore plan the run clock has reached, once, even when the report knows them too", () => {
    const planned = { startNs: 0, faults: [{ entity: "CT-1", fault: "fan_failure", atS: 240 }] };
    const open = blankReport({ mode: "explore", ended: false });
    const events = overlayOf([], open, whole, planned);
    expect(events.map((event) => [event.kind, event.atNs, event.label])).toEqual([
      ["fault", 240e9, "CT-1 fan failure"],
    ]);
    const known = {
      ...report,
      mode: "explore",
      ended: false,
      disturbances: [{ ...report.disturbances[1], onsetNs: 240e9 }],
    };
    expect(
      overlayOf([], known as typeof report, whole, planned).filter((event) => event.kind === "fault"),
    ).toHaveLength(1);
    expect(overlayOf([], null, { startNs: 0, endNs: 100e9 }, planned)).toEqual([]);
  });
});
