import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import { domain } from "../../../test/plantFixture";
import { defaultTrendSize, resizedTrend, savedTrendsOf, savedTrendsText, stillExists, trendSizes } from "./savedTrends";

const plant = parsePlant(domain);

describe("the size of the trends panel", () => {
  it("is taller by default once three or more channels are pinned", () => {
    expect([0, 1, 2].map(defaultTrendSize)).toEqual(["normal", "normal", "normal"]);
    expect([3, 5, 8].map(defaultTrendSize)).toEqual(["tall", "tall", "tall"]);
  });

  it("steps one size larger or smaller and stops at the ends", () => {
    expect(resizedTrend("normal", "larger")).toBe("tall");
    expect(resizedTrend("normal", "smaller")).toBe("compact");
    expect(resizedTrend("compact", "smaller")).toBe("compact");
    expect(resizedTrend(trendSizes[trendSizes.length - 1] ?? "normal", "larger")).toBe("full");
  });
});

describe("what a viewer keeps of its trends between page loads", () => {
  const saved = {
    open: true,
    size: "tall" as const,
    watched: [
      { entity: "CH-1", channel: "power" },
      { entity: "CT-1", channel: "basin_temp" },
    ],
  };

  it("round-trips through its text", () => {
    expect(savedTrendsOf(savedTrendsText(saved))).toEqual(saved);
  });

  it.each([
    ["nothing", null],
    ["text that is not JSON", "{nope"],
    ["JSON that is not an object", "[1]"],
    ["an object with no pins", '{"open":"yes"}'],
  ])("reads %s as nothing saved, or an empty list", (_name, text) => {
    const read = savedTrendsOf(text);
    expect(read === null || read.watched.length === 0).toBe(true);
  });

  it("drops what is not a pin, an unknown size, and more pins than the panel holds", () => {
    const many = Array.from({ length: 12 }, (_, index) => ({ entity: `E-${String(index)}`, channel: "c" }));
    const read = savedTrendsOf(
      JSON.stringify({ open: true, size: "huge", watched: [{ entity: 1 }, "x", null, ...many] }),
    );
    expect(read?.size).toBeNull();
    expect(read?.watched).toHaveLength(8);
    expect(read?.watched[0]).toEqual({ entity: "E-0", channel: "c" });
  });

  it("knows which pins still exist in the plant: the instance and its channel", () => {
    const first = plant.instances[0];
    const channel = plant.kinds.get(first?.kind ?? "")?.channels[0]?.name ?? "";
    expect(stillExists(plant, { entity: first?.id ?? "", channel })).toBe(true);
    expect(stillExists(plant, { entity: first?.id ?? "", channel: "no_such_channel" })).toBe(false);
    expect(stillExists(plant, { entity: "GONE-1", channel })).toBe(false);
  });
});
