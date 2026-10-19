import { describe, expect, it } from "vitest";
import { humanise, kindLabel, shortDigest, spaced } from "./words";

describe("shared words", () => {
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

  it("shortens a policy digest to its scheme and the first and last eight hex digits, and leaves anything else as it is", () => {
    const digest = `sha256:${"a".repeat(8)}${"0".repeat(48)}${"b".repeat(8)}`;
    expect(shortDigest(digest)).toBe("sha256:aaaaaaaa…bbbbbbbb");
    expect(shortDigest("sha256:abc")).toBe("sha256:abc");
    expect(shortDigest("")).toBe("");
  });
});
