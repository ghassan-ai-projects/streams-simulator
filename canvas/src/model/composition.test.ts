import { describe, expect, it } from "vitest";
import { compositionPlant, connectionsOf, equipmentOf, parseComposition, serialise } from "./composition";
import {
  addEquipment,
  connect,
  disconnect,
  issueTargets,
  moveInstance,
  newComposition,
  removeInstance,
  renameInstance,
} from "./compositionEdits";
import { parseLibrary, parsePresets } from "./library";

const doc = {
  id: "lab",
  title: "Lab",
  description: "base",
  site: { name: "Lab site" },
  areas: [{ name: "Site" }, { name: "Cooling", parent: "Site" }],
  profiles: [{ name: "nominal" }],
  equipment: [
    { use: "streamsim/chiller@1", names: ["CH-1", "CH-2"], area: "Cooling" },
    { use: "streamsim/cooling-tower@1", as: "tower", preset: "small", names: ["CT-1"], area: "Cooling" },
    { use: "streamsim/weather@1", id_template: "WX-{n}", area: "Site" },
  ],
  connections: [
    { from: "CH-1.heat_out", to: "CT-1.heat_in" },
    { from: "CT-1.basin_out", to: "CH-2.cond_in" },
  ],
};

describe("composition", () => {
  it("reads equipment, connections and a plant the graph can draw", () => {
    expect(equipmentOf(doc).map((entry) => entry.names)).toEqual([["CH-1", "CH-2"], ["CT-1"], ["WX-{n}"]]);
    expect(connectionsOf(doc)[1]).toEqual({ index: 1, from: "CT-1.basin_out", to: "CH-2.cond_in" });
    const plant = compositionPlant(doc, (component) => (component === "streamsim/chiller" ? "Chiller" : "?"));
    expect(plant.instances.map((instance) => [instance.id, instance.kind, instance.component])).toEqual([
      ["CH-1", "Chiller", "streamsim/chiller"],
      ["CH-2", "Chiller", "streamsim/chiller"],
      ["CT-1", "tower", "streamsim/cooling-tower"],
      ["WX-{n}", "?", "streamsim/weather"],
    ]);
    expect(plant.links[0]).toEqual({ from: "CH-1", to: "CT-1", type: "", fromPort: "heat_out", toPort: "heat_in" });
    expect(plant.areas[1]).toEqual({ name: "Cooling", parent: "Site" });
    expect(parseComposition("[1]")).toBeNull();
    expect(parseComposition("{nope")).toBeNull();
    expect(serialise({ a: 1 })).toBe('{\n  "a": 1\n}\n');
  });

  it("adds, renames, moves, connects and removes instances, keeping every other field", () => {
    const added = addEquipment(doc, { use: "streamsim/pump@1", preset: "", name: "P-1", area: "Site" });
    expect(equipmentOf(added).at(-1)).toMatchObject({ use: "streamsim/pump@1", names: ["P-1"], preset: "" });
    const wired = connect(added, "P-1.flow_out", "CH-1.flow_in");
    expect(connect(wired, "P-1.flow_out", "CH-1.flow_in")).toBe(wired);
    const renamed = renameInstance(wired, "CH-1", "CH-9");
    expect(equipmentOf(renamed)[0]?.names).toEqual(["CH-9", "CH-2"]);
    expect(connectionsOf(renamed).map((wire) => `${wire.from}>${wire.to}`)).toEqual([
      "CH-9.heat_out>CT-1.heat_in",
      "CT-1.basin_out>CH-2.cond_in",
      "P-1.flow_out>CH-9.flow_in",
    ]);
    const moved = moveInstance(renamed, "CH-2", "Site");
    expect(equipmentOf(moved).map((entry) => [entry.names, entry.area])).toContainEqual([["CH-2"], "Site"]);
    expect(equipmentOf(moveInstance(moved, "CT-1", "Site"))[1]?.area).toBe("Site");
    const removed = removeInstance(moved, "CH-9");
    expect(equipmentOf(removed).flatMap((entry) => entry.names)).not.toContain("CH-9");
    expect(connectionsOf(removed)).toHaveLength(1);
    expect(equipmentOf(removeInstance(doc, "WX-{n}"))).toHaveLength(2);
    expect(connectionsOf(disconnect(doc, 0))).toEqual([{ index: 0, from: "CT-1.basin_out", to: "CH-2.cond_in" }]);
    expect(removed.profiles).toBe(doc.profiles);
  });

  it("starts a new composition from a base without its equipment, links or description", () => {
    const fresh = newComposition(doc, "lab-two", "Lab two");
    expect(fresh).toMatchObject({ id: "lab-two", title: "Lab two", equipment: [], connections: [], areas: doc.areas });
    expect(fresh).not.toHaveProperty("description");
  });

  it("puts each issue of the binary on the instances its pointer names", () => {
    const targets = issueTargets(doc, [
      { pointer: "/equipment/0/area", message: "bad area" },
      { pointer: "/connections/1/from", message: "no port" },
      { pointer: "/equipment", message: "too few" },
    ]);
    expect(targets.instances.get("CH-1")).toEqual(["bad area"]);
    expect(targets.instances.get("CH-2")).toEqual(["bad area", "no port"]);
    expect(targets.instances.get("CT-1")).toEqual(["no port"]);
    expect(targets.listed.map((entry) => entry.target)).toEqual(["CH-1", "CT-1", null]);
  });

  it("lists the library's equipment and a component's presets", () => {
    const library = parseLibrary({
      components: [
        { id: "streamsim/pump", version: 2, kind: "equipment", title: "Pump" },
        { id: "streamsim/rtd", version: 1, kind: "instrument", title: "RTD" },
        { id: "streamsim/ahu", kind: "equipment", title: "Air-handling unit" },
      ],
    });
    expect(library).toEqual([
      { component: "streamsim/ahu", ref: "streamsim/ahu@1", title: "Air-handling unit" },
      { component: "streamsim/pump", ref: "streamsim/pump@2", title: "Pump" },
    ]);
    expect(parseLibrary(null)).toEqual([]);
    const presets = parsePresets({ document: { default_preset: "medium", presets: { small: { title: "Small" } } } });
    expect(presets).toEqual({ options: [{ name: "small", title: "Small" }], preferred: "medium" });
  });
});
