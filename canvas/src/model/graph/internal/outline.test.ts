import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import { belowCount, groups, marked, matchWords } from "./outline";

const plant = parsePlant({
  id: "outline",
  site: { name: "Site", areas: [{ name: "Site" }, { name: "Room", parent: "Site" }] },
  instances: [{ id: "P-1", kind: "pump", area: "Room" }],
});

describe("outline projection", () => {
  it("keeps empty parents and counts descendants while filtering by kind", () => {
    const shown = groups(plant, null, "", { emptyAreas: "show" });
    expect(shown.map((group) => group.area)).toEqual(["Site", "Room"]);
    expect(belowCount(shown, 0)).toBe(1);
    expect(belowCount(shown, 1)).toBe(0);
    expect(belowCount(shown, 20)).toBe(0);
    const filtered = groups(plant, null, "pump", { emptyAreas: "show" });
    expect(filtered).toHaveLength(1);
    expect(matchWords(filtered, "pump", false)).toBe("1 instance matches.");
  });

  it("projects binary issue marks without changing the plant", () => {
    const shown = groups(plant, null, "", { emptyAreas: "hide" });
    expect(marked(shown, undefined)).toBe(shown);
    expect(marked(shown, new Set())).toBe(shown);
    expect(marked(shown, new Set(["P-1"]))[0]?.items[0]?.flagged).toBe(true);
    expect(shown[0]?.items[0]?.flagged).toBeUndefined();
    expect(matchWords([], "missing", false)).toBe('No instance matches "missing".');
    expect(matchWords([], "", true)).toBe("No instances yet.");
  });

  it("moves an empty sub-area below every populated area, never between them", () => {
    const boiler = parsePlant({
      id: "boiler",
      site: {
        name: "Site",
        areas: [
          { name: "Site" },
          { name: "Boiler house", parent: "Site" },
          { name: "Gas train", parent: "Boiler house" },
          { name: "Distribution", parent: "Site" },
          { name: "Process hall", parent: "Site" },
          { name: "Heat users", parent: "Process hall" },
        ],
      },
      instances: [
        { kind: "meter", names: ["M-1", "M-2"], area: "Site" },
        { kind: "boiler", names: ["B-1", "B-2", "B-3", "B-4", "B-5", "B-6"], area: "Boiler house" },
        { kind: "pump", names: ["P-1", "P-2", "P-3"], area: "Distribution" },
        { kind: "pump", names: ["P-4", "P-5", "P-6"], area: "Process hall" },
      ],
    });
    const shown = groups(boiler, null, "", { emptyAreas: "show" });
    expect(shown.map((group) => [group.area, group.items.length])).toEqual([
      ["Site", 2],
      ["Boiler house", 6],
      ["Distribution", 3],
      ["Process hall", 3],
      ["Gas train", 0],
      ["Heat users", 0],
    ]);
  });
});
