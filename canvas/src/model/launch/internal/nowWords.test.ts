import { describe, expect, it } from "vitest";
import type { CallOutcome } from "./launch";
import { nowLines, outcomeWords } from "./nowWords";

const call: CallOutcome = {
  entity: "CH-1",
  effector: "chiller__set_setpoint",
  atS: 720,
  accepted: true,
  reason: "",
  effectEtaNs: null,
  immediate: false,
  atMs: 0,
};

describe("words for commands", () => {
  it("says what a planned call came to, and marks one given during the run as now", () => {
    expect(outcomeWords(call)).toBe("00:12:00 · CH-1 · set setpoint: accepted");
    expect(outcomeWords({ ...call, accepted: false, reason: "interlock_refused", immediate: true })).toBe(
      "00:12:00 · CH-1 · set setpoint: command not accepted: interlock_refused · now",
    );
  });

  it("lists the commands given during the run, calls and faults together, in the order of their time", () => {
    const planned = { ...call, atS: 10 };
    const later = { ...call, atS: 900, immediate: true };
    const fault = { entity: "CT-1", fault: "tower__fan_failure", atS: 800 };
    expect(nowLines([planned, later], [fault])).toEqual([
      "00:13:20 · CT-1 · fault injected: fan failure · now",
      "00:15:00 · CH-1 · set setpoint: accepted · now",
    ]);
    expect(nowLines([planned], [])).toEqual([]);
  });
});
