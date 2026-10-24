import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import { faultCategoryOf, faultGroups } from "./faultGroups";

describe("faults grouped by the layer they act on", () => {
  const categories = {
    fan_failure: "world",
    sensor_drift: "measurement",
    stuck_valve: "world_and_measurement",
    pump_wear: "world",
  } as const;

  it("groups by category in a fixed order and keeps the plant's order inside a group", () => {
    const faults = ["sensor_drift", "pump_wear", "stuck_valve", "fan_failure"];
    expect(faultGroups(faults, categories)).toEqual([
      { label: "Process (world)", faults: ["pump_wear", "fan_failure"] },
      { label: "Measurement (sensor)", faults: ["sensor_drift"] },
      { label: "World and measurement", faults: ["stuck_valve"] },
    ]);
  });

  it("puts a fault the binary did not categorise last, under Other faults", () => {
    expect(faultGroups(["fan_failure", "mystery"], categories)).toEqual([
      { label: "Process (world)", faults: ["fan_failure"] },
      { label: "Other faults", faults: ["mystery"] },
    ]);
  });

  it("leaves the list flat when nothing is categorised, or the categories are not there", () => {
    expect(faultGroups(["a", "b"], {})).toEqual([{ label: "", faults: ["a", "b"] }]);
    expect(faultGroups(["a", "b"], undefined)).toEqual([{ label: "", faults: ["a", "b"] }]);
    expect(faultGroups([], categories)).toEqual([{ label: "", faults: [] }]);
  });
});

describe("the category of one fault on one instance", () => {
  const plant = parsePlant({ id: "mini", instances: [{ kind: "tower", names: ["CT-1"] }], kinds: [] });
  const categories = { tower: { fan_failure: "world" } } as const;

  it("looks the fault up under the instance's kind", () => {
    expect(faultCategoryOf(plant, categories, "CT-1", "fan_failure")).toBe("world");
  });

  it("is unknown for an instance, a kind or a fault the catalog does not list", () => {
    expect(faultCategoryOf(plant, categories, "CT-9", "fan_failure")).toBeNull();
    expect(faultCategoryOf(plant, {}, "CT-1", "fan_failure")).toBeNull();
    expect(faultCategoryOf(plant, categories, "CT-1", "other")).toBeNull();
  });
});
