import { describe, expect, it } from "vitest";
import { compositionPlant } from "../../../test/modelFixtures";
import { linkRefusal } from "./linkRefusal";
import type { ComponentSpec } from "./ports";

const specs = new Map<string, ComponentSpec>([
  ["lib/pump", { category: "cool", ports: [{ name: "flow_out", type: "chilled_water", direction: "out" }] }],
  ["lib/tank", { category: "cool", ports: [{ name: "inflow_in", type: "chilled_water", direction: "in" }] }],
  ["lib/valve", { category: "cool", ports: [{ name: "pressure_in", type: "pressure", direction: "in" }] }],
  ["lib/any", { category: "cool", ports: [{ name: "any_in", type: "", direction: "in" }] }],
]);

const plant = compositionPlant(
  {
    equipment: [
      { use: "lib/pump@1", names: ["pump-1"] },
      { use: "lib/tank@1", names: ["tank-1"] },
      { use: "lib/valve@1", names: ["valve-1"] },
      { use: "lib/any@1", names: ["any-1"] },
    ],
  },
  (component) => component,
);

describe("linkRefusal", () => {
  it("accepts an output and an input of the same type, or an input that takes any type", () => {
    expect(linkRefusal(plant, specs, "pump-1.flow_out", "tank-1.inflow_in")).toBeNull();
    expect(linkRefusal(plant, specs, "pump-1.flow_out", "any-1.any_in")).toBeNull();
  });

  it("refuses an input of another type, naming both ends and both types", () => {
    expect(linkRefusal(plant, specs, "pump-1.flow_out", "valve-1.pressure_in")).toBe(
      "pump-1's flow out is chilled water, but valve-1's pressure in takes pressure. A link joins an output to an input of the same type.",
    );
  });

  it("leaves an end it cannot type to the binary", () => {
    expect(linkRefusal(plant, specs, "ghost-1.flow_out", "tank-1.inflow_in")).toBeNull();
    expect(linkRefusal(plant, specs, "pump-1.flow_out", "tank-1.missing_in")).toBeNull();
  });
});
