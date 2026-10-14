import { describe, expect, it } from "vitest";
import type { Composition } from "./composition";
import { equipmentOf } from "./composition";
import { addEquipment, moveInstance } from "./compositionEdits";

const empty = { id: "lab", title: "Lab", areas: [{ name: "Site" }, { name: "Hall" }], equipment: [] };

function add(doc: Composition, use: string, name: string, preset = "", area = "Site") {
  return addEquipment(doc, { use, preset, name, area });
}

function kinds(doc: Composition): readonly string[] {
  return equipmentOf(doc).map((entry) => entry.as || entry.use);
}

describe("several instances of one component", () => {
  it("put a second pump with the same size and area on the first pump's entry", () => {
    const doc = add(add(empty, "streamsim/pump@1", "pump-a"), "streamsim/pump@1", "pump-b");
    expect(equipmentOf(doc)).toHaveLength(1);
    expect(equipmentOf(doc)[0]?.names).toEqual(["pump-a", "pump-b"]);
  });

  it("give a chiller of another size its own kind name", () => {
    const first = add(empty, "streamsim/chiller@1", "ch-1", "medium");
    const doc = add(first, "streamsim/chiller@1", "ch-2", "large");
    expect(equipmentOf(doc).map((entry) => [entry.as, entry.preset, entry.names])).toEqual([
      ["", "medium", ["ch-1"]],
      ["chiller_2", "large", ["ch-2"]],
    ]);
    expect(new Set(kinds(doc)).size).toBe(2);
  });

  it("give an air-handling unit in another area its own kind name, then reuse it", () => {
    const base = add(empty, "streamsim/ahu@1", "ahu-1");
    const other = add(base, "streamsim/ahu@1", "ahu-2", "", "Hall");
    const third = add(other, "streamsim/ahu@1", "ahu-3", "", "Hall");
    expect(equipmentOf(third).map((entry) => [entry.as, entry.names, entry.area])).toEqual([
      ["", ["ahu-1"], "Site"],
      ["ahu_2", ["ahu-2", "ahu-3"], "Hall"],
    ]);
  });

  it("keep a unique kind name when a shared entry is split by moving one instance", () => {
    const doc = add(add(empty, "streamsim/pump@1", "p-1"), "streamsim/pump@1", "p-2");
    const moved = moveInstance(doc, "p-2", "Hall");
    expect(equipmentOf(moved).map((entry) => [entry.as, entry.names, entry.area])).toEqual([
      ["", ["p-1"], "Site"],
      ["pump_2", ["p-2"], "Hall"],
    ]);
    const back = moveInstance(moved, "p-2", "Site");
    expect(equipmentOf(back).map((entry) => [entry.names, entry.area])).toEqual([[["p-1", "p-2"], "Site"]]);
  });

  it("never reuse a taken kind name", () => {
    const doc = {
      ...empty,
      equipment: [
        { use: "x/pump@1", names: ["a"], area: "Site" },
        { use: "x/pump@1", as: "pump_2", preset: "big", names: ["b"], area: "Site" },
      ],
    };
    const next = add(doc, "x/pump@1", "c", "huge");
    expect(kinds(next)).toEqual(["x/pump@1", "pump_2", "pump_3"]);
  });
});
