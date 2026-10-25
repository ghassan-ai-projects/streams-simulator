import { describe, expect, it } from "vitest";
import { component } from "./decoders";

const entry = {
  id: "streamsim/air-compressor",
  title: "Compressor",
  class: "rotating/compressor/air_compressor",
  tags: [],
};

describe("the parameters and sizes of an equipment", () => {
  it("reads a numeric parameter with its unit, bounds and default, and the sizes with their default", () => {
    const spec = component({
      entry,
      document: {
        ports: [],
        default_preset: "medium",
        presets: { large: { title: "Large (250 kW)", values: {} }, medium: { values: {} } },
        parameters: [
          {
            name: "load_kpa",
            title: "Header pressure",
            type: "number",
            unit: "kPa",
            min: 550,
            max: 750,
            default: 660,
            targets: ["a"],
          },
        ],
      },
    });
    expect(spec).toMatchObject({
      parameters: [
        {
          name: "load_kpa",
          title: "Header pressure",
          unit: "kPa",
          type: "number",
          min: 550,
          max: 750,
          default: 660,
          values: [],
        },
      ],
      presets: [
        { name: "large", title: "Large (250 kW)" },
        { name: "medium", title: "medium" },
      ],
      defaultPreset: "medium",
    });
  });

  it("reads an enumerated parameter with its values, and one without a unit, bounds or title", () => {
    const spec = component({
      entry,
      document: {
        ports: [],
        parameters: [
          {
            name: "characteristic",
            title: "Flow characteristic",
            type: "enum",
            values: ["linear", "equal_percentage"],
            default: "linear",
            cases: {},
          },
          { name: "n", type: "integer" },
        ],
      },
    });
    expect(spec).toMatchObject({
      parameters: [
        {
          name: "characteristic",
          type: "enum",
          values: ["linear", "equal_percentage"],
          default: "linear",
          min: null,
          max: null,
        },
        { name: "n", title: "n", unit: "", type: "integer", default: null },
      ],
    });
  });

  it("skips a parameter it cannot describe and gives none when there are none", () => {
    const odd = component({
      entry,
      document: {
        ports: [],
        parameters: [{ name: "a", type: "odd" }, { type: "number" }, 3, { name: "b", type: "number", min: "x" }],
      },
    });
    expect(odd?.parameters).toMatchObject([{ name: "b", min: null }]);
    const bare = component({ entry, document: { ports: [], presets: 4 } });
    expect(bare).toEqual({ ports: [], category: "env", title: "Compressor" });
  });
});
