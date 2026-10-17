import { describe, expect, it } from "vitest";
import {
  compositionPlant,
  connectionsOf,
  equipmentOf,
  linkEnd,
  parseComposition,
  serialise,
  shortKind,
  suggestedName,
} from "./composition";
import {
  addEquipment,
  connect,
  disconnect,
  issueTargets,
  isTemplate,
  moveInstance,
  removeInstance,
  renameInstance,
} from "./compositionEdits";
import type { LibraryEntry } from "../../plant";
import { searchLibrary } from "../../plant";

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

  it("puts each issue of the binary on the instances its pointer names", () => {
    const targets = issueTargets(doc, [
      { pointer: "/equipment/0/area", message: "bad area" },
      { pointer: "/connections/1/from", message: "no port" },
      { pointer: "/equipment", message: "too few" },
      { pointer: "/equipment/0/names/1", message: "bad name" },
      { pointer: "/equipment/0/names/7", message: "gone" },
    ]);
    expect(targets.instances.get("CH-1")).toEqual(["bad area"]);
    expect(targets.instances.get("CH-2")).toEqual(["bad area", "no port", "bad name"]);
    expect(targets.instances.get("CT-1")).toEqual(["no port"]);
    expect(targets.listed.map((entry) => entry.target)).toEqual(["CH-1", "CT-1", null, "CH-2", null]);
    expect([isTemplate(doc, "WX-{n}"), isTemplate(doc, "CH-1")]).toEqual([true, false]);
  });

  it("searches the library's equipment by its words", () => {
    const library: readonly LibraryEntry[] = [
      {
        component: "streamsim/pump",
        ref: "streamsim/pump@2",
        title: "Pump",
        words: "pump streamsim/pump rotating/pump water",
      },
      {
        component: "streamsim/ahu",
        ref: "streamsim/ahu@1",
        title: "Air-handling unit",
        words: "air-handling unit streamsim/ahu",
      },
    ];
    expect(searchLibrary(library, "")).toHaveLength(2);
    expect(searchLibrary(library, "AIR unit").map((entry) => entry.title)).toEqual(["Air-handling unit"]);
    expect(searchLibrary(library, "pump chiller")).toEqual([]);
  });

  it("suggests the next free name like its siblings, or from the title's initials", () => {
    expect(suggestedName(doc, "streamsim/chiller@1", "Chiller")).toBe("CH-3");
    expect(suggestedName(doc, "streamsim/pump@1", "Chilled water pump")).toBe("CWP-1");
    expect(suggestedName(doc, "streamsim/cooling-tower@1", "Cooling tower")).toBe("CT-2");
    expect(suggestedName({}, "x@1", "")).toBe("X-1");
    const longPump = {
      equipment: [{ use: "streamsim/pump@1", names: ["pump-house-secondary-circuit-feed-north-07"] }],
    };
    expect(suggestedName(longPump, "streamsim/pump@1", "Centrifugal pump")).toBe("CP-1");
  });

  it("says a link from one instance's side in words", () => {
    const [first] = connectionsOf(doc);
    if (first === undefined) throw new Error("no link");
    expect(linkEnd(first, "CH-1")).toEqual({ outgoing: true, port: "heat out", other: "CT-1", otherPort: "heat in" });
    expect(linkEnd(first, "CT-1")).toEqual({ outgoing: false, port: "heat in", other: "CH-1", otherPort: "heat out" });
  });

  it("names a kind the way the plant screen does, from the component's last segment", () => {
    expect([shortKind("streamsim/air-receiver"), shortKind("chiller"), shortKind("lib/a/b-c-d")]).toEqual([
      "air_receiver",
      "chiller",
      "b_c_d",
    ]);
  });
});
