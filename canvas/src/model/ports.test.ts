import { describe, expect, it } from "vitest";
import { chillerPorts, composition, domain } from "../test/plantFixture";
import { parsePlant } from "./plant";
import { familyOf, parseComponent, parsePorts, pinsOf, portType, portsMatching, shapeOf } from "./ports";

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

  it("names a component's category from its library class, then its tags", () => {
    const of = (componentClass: string, tags: string[]) =>
      parseComponent({ entry: { class: componentClass, tags }, document: chillerPorts }).category;
    expect(of("electrical/switchgear/feeder_meter", ["hvac"])).toBe("elec");
    expect(of("process/production/line", [])).toBe("prod");
    expect(of("environment/site/weather", [])).toBe("env");
    expect(of("rotating/compressor/air", ["compressed-air"])).toBe("air");
    expect(of("rotating/pump/centrifugal", ["fluids"])).toBe("cool");
    expect(of("rotating/conveyor/belt", ["drives"])).toBe("prod");
    expect(of("other/thing", [])).toBe("env");
    expect(parseComponent({ entry: { class: "x", tags: [1, "hvac"] }, document: chillerPorts }).ports).toHaveLength(4);
  });

  it("keeps declared component title and description with a document-title fallback", () => {
    const declared = parseComponent({
      entry: { title: "Summary title", class: "x" },
      document: { title: "Document title", description: "Declared purpose." },
    });
    expect([declared.title, declared.description]).toEqual(["Summary title", "Declared purpose."]);
    expect(parseComponent({ entry: {}, document: { title: "Document title" } }).title).toBe("Document title");
  });

  it("ignores absent and malformed component metadata", () => {
    expect(parseComponent(null)).toEqual({ category: "env", ports: [] });
    expect(parseComponent({ entry: null, document: { title: 12, description: {}, ports: "bad" } })).toEqual({
      category: "env",
      ports: [],
    });
    expect(parseComponent({ entry: { title: "  " }, document: { title: "", description: " " } })).toEqual({
      category: "env",
      ports: [],
    });
  });

  it("marks the pins a link uses as connected", () => {
    const plant = parsePlant(domain, composition);
    const chiller = plant.instances[0];
    if (chiller === undefined) throw new Error("fixture");
    const pins = pinsOf(plant, chiller, parsePorts(chillerPorts));
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
    const pins = pinsOf(plant, chiller, parsePorts(chillerPorts));
    expect(pins.ins.map((pin) => pin.connection)).toEqual(["unknown", "unknown"]);
    expect(pins.outs.map((pin) => pin.connection)).toEqual(["unknown", "unknown"]);
  });

  it("falls back to the ports the links name, untyped, when the component is unknown", () => {
    const plant = parsePlant(domain, composition);
    const meter = plant.instances[2];
    if (meter === undefined) throw new Error("fixture");
    expect(pinsOf(plant, meter, undefined)).toEqual({
      ins: [{ name: "loads_in", type: "", family: null, connection: "connected" }],
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
});
