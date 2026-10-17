import { describe, expect, it } from "vitest";
import { domain } from "../../../test/plantFixture";
import { parsePlant } from "../../plant";
import { areaCounts } from "./areas";

describe("areas", () => {
  it("counts areas in plant order, matching the outline, not by count", () => {
    const plant = parsePlant({
      id: "plant",
      site: {
        name: "Site",
        areas: [{ name: "Site" }, { name: "Hall 1", parent: "Site" }, { name: "Utilities", parent: "Site" }],
      },
      instances: [
        { kind: "meter", names: ["U-1", "U-2", "U-3"], area: "Utilities" },
        { kind: "pump", names: ["P-1"], area: "Hall 1" },
        { kind: "pump", names: ["P-2"], area: "Yard" },
      ],
    });
    expect(areaCounts(plant)).toEqual([
      ["Hall 1", 1],
      ["Utilities", 3],
      ["Yard", 1],
    ]);
    expect(areaCounts(parsePlant(domain))).toEqual([
      ["Cooling", 2],
      ["Power", 1],
    ]);
  });
});
