import { describe, expect, it } from "vitest";
import type { Composition } from "./composition";
import { areaInUse, areaNameProblem, removeArea, renameArea } from "./areaEdits";

const doc: Composition = {
  id: "lab",
  areas: [{ name: "Site" }, { name: "Cooling", parent: "Site" }, { name: "Yard", parent: "Site" }],
  equipment: [
    { use: "lib/chiller@1", names: ["CH-1"], area: "Cooling" },
    { use: "lib/tower@1", id_template: "CT-{n}", count: { default: 3 }, area: "Cooling" },
    { use: "lib/pump@1", names: ["P-1"], area: "Yard" },
  ],
};

describe("area edits", () => {
  it("renames an area in the area list, in child parents and in equipment", () => {
    const next = renameArea(doc, "Site", "Campus");
    expect(next.areas).toEqual([
      { name: "Campus" },
      { name: "Cooling", parent: "Campus" },
      { name: "Yard", parent: "Campus" },
    ]);
    expect(renameArea(doc, "Yard", "Dock").equipment).toEqual([
      { use: "lib/chiller@1", names: ["CH-1"], area: "Cooling" },
      { use: "lib/tower@1", id_template: "CT-{n}", count: { default: 3 }, area: "Cooling" },
      { use: "lib/pump@1", names: ["P-1"], area: "Dock" },
    ]);
  });

  it("refuses a rename to an empty, taken or unchanged name", () => {
    expect(renameArea(doc, "Yard", "Cooling")).toBe(doc);
    expect(renameArea(doc, "Yard", "")).toBe(doc);
    expect(renameArea(doc, "Yard", "Yard")).toBe(doc);
    expect(areaNameProblem(doc, "Cooling", "Yard")).toBe('There is already an area named "Cooling".');
    expect(areaNameProblem(doc, "Yard", "Yard")).toBeNull();
    expect(areaNameProblem(doc, "", null)).toBe("Give the area a name.");
  });

  it("says why an area cannot be deleted, counting instances and areas inside it", () => {
    expect(areaInUse(doc, "Cooling")).toBe("Move the 4 instances out of it first.");
    expect(areaInUse(doc, "Site")).toBe("Delete the 2 areas inside it first.");
    expect(areaInUse({ ...doc, equipment: [{ use: "lib/pump@1", names: ["P-1"], area: "Yard" }] }, "Yard")).toBe(
      "Move the 1 instance out of it first.",
    );
    expect(areaInUse({ ...doc, equipment: [], areas: [{ name: "Yard" }] }, "Yard")).toBeNull();
  });

  it("removes an empty area and leaves an area in use", () => {
    const moved = { ...doc, equipment: [{ use: "lib/pump@1", names: ["P-1"], area: "Cooling" }] };
    expect(removeArea(doc, "Cooling")).toBe(doc);
    expect(removeArea(moved, "Yard")).toMatchObject({ areas: [{ name: "Site" }, { name: "Cooling", parent: "Site" }] });
  });
});
