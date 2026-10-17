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
});
