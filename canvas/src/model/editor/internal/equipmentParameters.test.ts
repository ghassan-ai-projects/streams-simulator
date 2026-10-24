import { describe, expect, it } from "vitest";
import type { Composition } from "./composition";
import { equipmentOf } from "./composition";
import { fieldIssues, setParameter, setPreset } from "./equipmentEdits";

const plant: Composition = {
  id: "lab",
  title: "Lab",
  areas: [{ name: "Site" }],
  equipment: [
    { use: "streamsim/pump@1", as: "chw_pump", names: ["P-1", "P-2", "P-3"], area: "Site", extra: true },
    { use: "streamsim/chiller@1", preset: "large", parameters: { setpoint_c: 6 }, names: ["CH-1"], area: "Site" },
    { use: "streamsim/pump@1", as: "cw_pump", names: ["P-4"], area: "Site" },
  ],
  connections: [{ from: "P-2.out", to: "CH-1.in" }],
};

const rows = (doc: Composition) =>
  equipmentOf(doc).map((entry) => [entry.as, entry.names, entry.preset, entry.parameters]);

describe("the preset and parameters of an instance", () => {
  it("reads them from the equipment entry that holds the instance", () => {
    expect(equipmentOf(plant)[1]).toMatchObject({ preset: "large", parameters: { setpoint_c: 6 } });
    expect(equipmentOf(plant)[0]?.parameters).toEqual({});
  });

  it("changes them in place on an entry that holds only that instance, keeping every other field", () => {
    const doc = setParameter(plant, "CH-1", "setpoint_c", 8);
    expect(equipmentOf(doc)[1]?.parameters).toEqual({ setpoint_c: 8 });
    expect(doc.connections).toBe(plant.connections);
    expect(equipmentOf(setPreset(plant, "CH-1", "small"))[1]).toMatchObject({
      preset: "small",
      parameters: { setpoint_c: 6 },
    });
  });

  it("gives an instance that shared an entry its own entry and kind name, and leaves the others as they were", () => {
    const doc = setParameter(plant, "P-2", "head_kpa", 120);
    expect(rows(doc)).toEqual([
      ["chw_pump", ["P-1", "P-3"], "", {}],
      ["", ["CH-1"], "large", { setpoint_c: 6 }],
      ["cw_pump", ["P-4"], "", {}],
      ["pump_2", ["P-2"], "", { head_kpa: 120 }],
    ]);
    expect(doc.equipment).toHaveLength(4);
    expect((doc.equipment as readonly Record<string, unknown>[])[3]).toMatchObject({ extra: true, area: "Site" });
  });

  it("puts an instance back on the entry that already has its preset and parameters", () => {
    const apart = setParameter(plant, "P-2", "head_kpa", 120);
    const back = setParameter(apart, "P-2", "head_kpa", null);
    expect(rows(back).map(([, names]) => names)).toEqual([["P-1", "P-3", "P-2"], ["CH-1"], ["P-4"]]);
  });

  it("removes the preset or a parameter that is cleared, and the parameters member when it is empty", () => {
    expect(equipmentOf(setPreset(plant, "CH-1", ""))[1]?.preset).toBe("");
    const cleared = setParameter(plant, "CH-1", "setpoint_c", null);
    expect("parameters" in ((cleared.equipment as readonly object[])[1] ?? {})).toBe(false);
  });

  it("keeps a parameter that stays when another is cleared", () => {
    const two = setParameter(plant, "CH-1", "flow", 3);
    expect(equipmentOf(setParameter(two, "CH-1", "flow", null))[1]?.parameters).toEqual({ setpoint_c: 6 });
  });

  it("changes nothing for an instance the composition does not hold", () => {
    expect(setParameter(plant, "X-9", "a", 1)).toBe(plant);
    expect(setPreset(plant, "X-9", "small")).toBe(plant);
  });
});

describe("the refusals of the binary that belong to one field", () => {
  const listed = [
    {
      target: "CH-1",
      message: '7 is outside the bounds of parameter "setpoint_c"',
      pointer: "/equipment/1/parameters/setpoint_c",
    },
    { target: "CH-1", message: 'chiller has no preset "huge"', pointer: "/equipment/1/preset" },
    { target: "P-1", message: "unrelated", pointer: "/equipment/0/names/0" },
  ];

  it("picks the ones at a parameter or at the preset of an entry", () => {
    expect(fieldIssues(listed, 1, "setpoint_c")).toEqual(['7 is outside the bounds of parameter "setpoint_c"']);
    expect(fieldIssues(listed, 1, "preset")).toEqual(['chiller has no preset "huge"']);
    expect(fieldIssues(listed, 0, "setpoint_c")).toEqual([]);
    expect(fieldIssues(listed, -1, "preset")).toEqual([]);
  });
});
