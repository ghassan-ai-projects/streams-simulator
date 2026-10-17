import { describe, expect, it } from "vitest";
import { chillerPorts, composition, domain } from "../../../test/plantFixture";
import { compositionPlant } from "../../../test/modelBridge";
import { parsePlant } from "./plant";
import type { ComponentSpec } from "./ports";
import { familyOf, noTargetReason, pinsOf, portType, portsMatching, shapeOf } from "./ports";

describe("ports and pins", () => {
  it("maps every library port type to one of nine families with a shape", () => {
    expect(familyOf("chilled_water")).toBe("water");
    expect(familyOf("wet_bulb_temp")).toBe("env");
    expect(familyOf("unknown")).toBeNull();
    expect(shapeOf("water")).toBe("circle");
    expect(shapeOf("power")).toBe("square");
    expect(shapeOf("pressure")).toBe("diamond");
  });

  it("separates look-alike hues by mark, so hue is never the only carrier", () => {
    expect(shapeOf("temp")).toBe("hexagon");
    expect(shapeOf("signal")).toBe("hexagon");
    expect(shapeOf("env")).toBe("bar");
    expect(shapeOf("heat")).toBe("bar");
    expect(shapeOf("temp")).not.toBe(shapeOf("heat"));
    expect(shapeOf("env")).not.toBe(shapeOf("signal"));
  });

  it("marks the pins a link uses as connected", () => {
    const plant = parsePlant(domain, composition);
    const chiller = plant.instances[0];
    if (chiller === undefined) throw new Error("fixture");
    const pins = pinsOf(plant, chiller, chillerPorts.ports);
    expect(pins.ins.map((pin) => [pin.name, pin.connection, pin.family])).toEqual([
      ["cond_in", "connected", "temp"],
      ["spare_in", "open", "water"],
    ]);
    expect(pins.outs.map((pin) => pin.connection)).toEqual(["connected", "connected"]);
  });

  it("marks a component's pins unknown when its links do not name ports", () => {
    const plant = parsePlant(domain);
    const chiller = plant.instances[0];
    if (chiller === undefined) throw new Error("fixture");
    const pins = pinsOf(plant, chiller, chillerPorts.ports);
    expect(pins.ins.map((pin) => pin.connection)).toEqual(["unknown", "unknown"]);
    expect(pins.outs.map((pin) => pin.connection)).toEqual(["unknown", "unknown"]);
  });

  it("falls back to the ports the links name, untyped, when the component is unknown", () => {
    const plant = parsePlant(domain, composition);
    const meter = plant.instances[2];
    if (meter === undefined) throw new Error("fixture");
    expect(pinsOf(plant, meter, undefined)).toEqual({
      ins: [{ name: "loads_in", type: "", family: null, connection: "connected", channel: null }],
      outs: [],
    });
    expect(pinsOf(parsePlant(domain), meter, undefined)).toEqual({ ins: [], outs: [] });
  });

  it("offers only ports that take what an output gives", () => {
    const spec = {
      category: "cool" as const,
      ports: [
        { name: "heat_in", type: "heat_rejection", direction: "in" as const },
        { name: "power_in", type: "electrical", direction: "in" as const },
        { name: "any_in", type: "", direction: "in" as const },
        { name: "power_out", type: "electrical", direction: "out" as const },
      ],
    };
    expect(portsMatching(spec, "in", "electrical")).toEqual(["power_in", "any_in"]);
    expect(portsMatching(spec, "in")).toEqual(["heat_in", "power_in", "any_in"]);
    expect(portsMatching(undefined, "out")).toEqual([]);
    expect([portType(spec, "power_out"), portType(spec, "nothing")]).toEqual(["electrical", ""]);
  });

  it("says why no instance takes an output, naming its type and the input types the plant has", () => {
    const specs = new Map<string, ComponentSpec>([
      ["lib/pump", { category: "cool", ports: [{ name: "flow_out", type: "chilled_water", direction: "out" }] }],
      ["lib/tank", { category: "cool", ports: [{ name: "inflow_in", type: "liquid_flow", direction: "in" }] }],
      ["lib/valve", { category: "cool", ports: [{ name: "pressure_in", type: "pressure", direction: "in" }] }],
    ]);
    const plant = compositionPlant(
      {
        equipment: [
          { use: "lib/pump@1", names: ["pump-1"] },
          { use: "lib/tank@1", names: ["tank-1"] },
          { use: "lib/valve@1", names: ["valve-1"] },
        ],
      },
      (component) => component,
    );
    const pump = plant.instances[0];
    if (pump === undefined) throw new Error("no pump");
    expect(noTargetReason(plant, specs, pump, "flow_out")).toBe(
      "No instance takes chilled water: pump-1's flow out is chilled water. The inputs in this plant take liquid flow and pressure.",
    );
    const alone = compositionPlant({ equipment: [{ use: "lib/pump@1", names: ["pump-1"] }] }, (c) => c);
    expect(noTargetReason(alone, specs, pump, "flow_out")).toMatch(/The other instances here have no inputs\.$/);
  });
});
