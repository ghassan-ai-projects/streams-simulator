import { describe, expect, it } from "vitest";
import { composition, domain } from "../test/plantFixture";
import { areaOf, linksOf, parsePlant } from "./plant";

describe("plant from a domain file", () => {
  it("reads instances, kinds, areas and links", () => {
    const plant = parsePlant(domain);
    expect(plant.id).toBe("mini");
    expect(plant.site).toBe("Mini");
    expect(plant.areas).toEqual([
      { name: "Mini", parent: null },
      { name: "Cooling", parent: "Mini" },
      { name: "Power", parent: "Mini" },
    ]);
    expect(plant.instances.map((instance) => instance.id)).toEqual(["CH-1", "CT-1", "MB-1"]);
    expect(plant.instances[2]?.component).toBeNull();
    expect(plant.kinds.get("chiller")).toEqual({
      name: "chiller",
      channels: [{ name: "chiller.power", unit: "kW" }],
      states: ["cop"],
      faults: ["chiller__fouling"],
      effectors: ["start"],
    });
    expect(plant.links[0]).toEqual({
      from: "CH-1",
      to: "CT-1",
      type: "tower__heat_in__chiller",
      fromPort: null,
      toPort: null,
    });
  });

  it("takes link ports from the composition when there is one", () => {
    const plant = parsePlant(domain, composition);
    expect(plant.links.map((link) => [link.fromPort, link.toPort])).toEqual([
      ["heat_out", "heat_in"],
      ["basin_out", "cond_in"],
      ["power_out", "loads_in"],
    ]);
  });

  it("tolerates a file that is not a domain", () => {
    const plant = parsePlant("nonsense");
    expect(plant.instances).toEqual([]);
    expect(plant.links).toEqual([]);
    expect(parsePlant({ instances: [{ id: "solo", kind: "k" }] }).instances).toEqual([
      { id: "solo", kind: "k", area: "", component: null },
    ]);
  });

  it("names an instance's area, falling back to the site, and lists its links", () => {
    const plant = parsePlant(domain);
    const meter = plant.instances[2];
    expect(meter && areaOf(plant, meter)).toBe("Power");
    expect(areaOf(plant, { id: "x", kind: "k", area: "", component: null })).toBe("Mini");
    expect(areaOf(parsePlant({}), { id: "x", kind: "k", area: "", component: null })).toBe("Plant");
    expect(linksOf(plant, "CT-1").incoming.map((link) => link.from)).toEqual(["CH-1"]);
    expect(linksOf(plant, "CT-1").outgoing.map((link) => link.to)).toEqual(["CH-1"]);
  });
});
