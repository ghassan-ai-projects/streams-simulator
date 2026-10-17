import { describe, expect, it } from "vitest";
import { failureLine, humanise, kindLabel, spaced } from "./words";

describe("shared words", () => {
  it("puts the display code after the diagnostic message", () => {
    expect(failureLine({ message: "The bridge did not answer.", displayCode: "unreachable" })).toBe(
      "The bridge did not answer. (unreachable)",
    );
  });

  it("spaces underscores and humanises ids", () => {
    expect(spaced("flow_rate_high")).toBe("flow rate high");
    expect(humanise("plant__flow_rate")).toBe("flow rate");
  });

  it("turns library ids into words and keeps the instrument", () => {
    expect(humanise("compressor_lead__oil_cooler_fouling")).toBe("oil cooler fouling");
    expect(humanise("cooling_tower__basin_temp__offset")).toBe("basin temp · offset");
    expect(humanise("discharge_temp")).toBe("discharge temp");
    expect(kindLabel("air_compressor_2")).toBe("air compressor");
    expect(kindLabel("compressor_lead__air_compressor_2")).toBe("air compressor");
    expect(kindLabel("chiller")).toBe("chiller");
  });
});
