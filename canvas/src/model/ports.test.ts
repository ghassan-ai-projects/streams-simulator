import { describe, expect, it } from "vitest";
import { chillerPorts, composition, domain } from "../test/plantFixture";
import { parsePlant } from "./plant";
import { familyOf, parsePorts, pinsOf, shapeOf } from "./ports";

describe("ports and pins", () => {
  it("maps every library port type to one of nine families with a shape", () => {
    expect(familyOf("chilled_water")).toBe("water");
    expect(familyOf("wet_bulb_temp")).toBe("env");
    expect(familyOf("unknown")).toBeNull();
    expect(shapeOf("water")).toBe("circle");
    expect(shapeOf("power")).toBe("square");
    expect(shapeOf("signal")).toBe("diamond");
  });

  it("reads a component document's ports and their directions", () => {
    expect(parsePorts(chillerPorts).map((port) => `${port.direction}:${port.name}`)).toEqual([
      "in:cond_in",
      "in:spare_in",
      "out:heat_out",
      "out:power_out",
    ]);
    expect(parsePorts(null)).toEqual([]);
  });

  it("marks the pins a link uses as connected", () => {
    const plant = parsePlant(domain, composition);
    const chiller = plant.instances[0];
    if (chiller === undefined) throw new Error("fixture");
    const pins = pinsOf(plant, chiller, parsePorts(chillerPorts));
    expect(pins.ins.map((pin) => [pin.name, pin.connected, pin.family])).toEqual([
      ["cond_in", true, "temp"],
      ["spare_in", false, "water"],
    ]);
    expect(pins.outs.map((pin) => pin.connected)).toEqual([true, true]);
  });

  it("falls back to the ports the links name, untyped, when the component is unknown", () => {
    const plant = parsePlant(domain, composition);
    const meter = plant.instances[2];
    if (meter === undefined) throw new Error("fixture");
    expect(pinsOf(plant, meter, undefined)).toEqual({
      ins: [{ name: "loads_in", type: "", family: null, connected: true }],
      outs: [],
    });
    expect(pinsOf(parsePlant(domain), meter, undefined)).toEqual({ ins: [], outs: [] });
  });
});
