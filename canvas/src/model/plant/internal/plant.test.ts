import { describe, expect, it } from "vitest";
import { composition, domain } from "../../../test/plantFixture";
import { areaOf, channelOf, effectorOf, linksOf, parsePlant, stateUnit } from "./plant";

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
      channels: [{ name: "chiller.power", unit: "kW", states: [], range: null }],
      states: ["cop"],
      units: new Map([["cop", ""]]),
      faults: ["chiller__fouling"],
      effectors: ["start"],
      effectorSpecs: [{ name: "start", entityArg: null, args: [] }],
    });
    expect(plant.links[0]).toEqual({
      from: "CH-1",
      to: "CT-1",
      type: "tower__heat_in__chiller",
      fromPort: null,
      toPort: null,
    });
  });

  it("finds the effector a kind declares for an instance", () => {
    const plant = parsePlant(domain);
    expect(effectorOf(plant, "CH-1", "start")).toEqual({ name: "start", entityArg: null, args: [] });
    expect(effectorOf(plant, "CH-1", "stop")).toBeUndefined();
    expect(effectorOf(plant, "nobody", "start")).toBeUndefined();
  });

  it("reads link ports and instruments the domain records, and units and states for KPIs", () => {
    const plant = parsePlant({
      ...domain,
      instances: [
        {
          kind: "chiller",
          names: ["CH-1"],
          area: "Cooling",
          provenance: { instruments: { cop_point: "lib/annunciator@1" } },
        },
      ],
      links: [{ from: "CH-1", to: "MB-1", type: "t", from_port: "power_out", to_port: "loads_in" }],
      kinds: [
        {
          name: "chiller",
          state: [{ name: "cop", unit: "ratio" }],
          channels: [
            {
              name: "chiller.alarm_state",
              range: { min: 0, max: 4 },
              states: [
                { value: 1, name: "unacknowledged", attention: "alarm" },
                { value: "x", name: "bad" },
                { value: 4, name: "shelved", attention: "loud" },
              ],
            },
          ],
        },
      ],
    });
    expect(plant.links[0]).toMatchObject({ fromPort: "power_out", toPort: "loads_in" });
    expect(channelOf(plant, "CH-1", "chiller.alarm_state")?.states).toEqual([
      { value: 1, name: "unacknowledged", attention: "alarm" },
      { value: 4, name: "shelved", attention: null },
    ]);
    expect(channelOf(plant, "CH-1", "chiller.alarm_state")?.range).toEqual({ min: 0, max: 4 });
    expect(channelOf(plant, "CH-1", "chiller.power")).toBeNull();
    expect(channelOf(plant, "nobody", "chiller.alarm_state")).toBeNull();
    expect(stateUnit(plant, "CH-1", "cop")).toBe("ratio");
    expect(stateUnit(plant, "nobody", "cop")).toBe("");
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
      { id: "solo", kind: "k", area: "", component: null, instruments: new Map() },
    ]);
  });

  it("names an instance's area, falling back to the site, and lists its links", () => {
    const plant = parsePlant(domain);
    const meter = plant.instances[2];
    expect(meter && areaOf(plant, meter)).toBe("Power");
    expect(areaOf(plant, { id: "x", kind: "k", area: "", component: null, instruments: new Map() })).toBe("Mini");
    expect(areaOf(parsePlant({}), { id: "x", kind: "k", area: "", component: null, instruments: new Map() })).toBe(
      "Plant",
    );
    expect(linksOf(plant, "CT-1").incoming.map((link) => link.from)).toEqual(["CH-1"]);
    expect(linksOf(plant, "CT-1").outgoing.map((link) => link.to)).toEqual(["CH-1"]);
  });
});
