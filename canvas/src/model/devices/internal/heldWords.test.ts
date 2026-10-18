import { describe, expect, it } from "vitest";
import { heldWords } from "./heldWords";

const unit = { pointer: "/metrics/0/unit", message: "needs a unit", keyword: "twin_unit_required", member: "unit" };
const resolution = {
  pointer: "/metrics/0/resolution",
  message: 'needs "resolution"',
  keyword: "twin_member_required",
  member: "resolution",
};

describe("saying why a device is not ready for a twin", () => {
  it("names the device and what it still needs, once each, in plain words", () => {
    expect(heldWords("pump_monitor", [unit])).toBe("pump_monitor still needs a unit");
    expect(heldWords("pump_monitor", [unit, resolution])).toBe("pump_monitor still needs a unit and a resolution");
    expect(heldWords("pump_monitor", [unit, unit])).toBe("pump_monitor still needs a unit");
    const range = { ...resolution, member: "eu_range" };
    expect(heldWords("pump", [unit, resolution, range])).toBe("pump still needs a unit, a resolution and a range");
    expect(heldWords("pump", [{ ...resolution, member: "odd_setting" }])).toBe(
      "pump still needs the odd setting setting",
    );
    expect(heldWords("pump", [{ ...resolution, member: "" }])).toBe("pump still needs more detail");
  });

  it("falls back to the first issue when the device lacks something else than a twin member", () => {
    const pattern = { pointer: "/kind", message: "bad", keyword: "pattern", member: "" };
    expect(heldWords("pump", [pattern])).toBe("pump is not ready for a twin: This value is not in the accepted form");
    expect(heldWords("pump", [unit, pattern])).toContain("pump is not ready for a twin");
    expect(heldWords("pump", [])).toBe("pump is not ready for a twin");
  });
});
