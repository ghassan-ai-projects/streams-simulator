import { describe, expect, it } from "vitest";
import {
  boundsWithTitles,
  capScale,
  countWords,
  headerReserve,
  splitName,
  titleBandHeight,
  titleScale,
} from "./frameTitle";

describe("frame title", () => {
  it("keeps the last word of a name apart so a cut name still tells its sisters apart", () => {
    expect(splitName("Production hall A")).toEqual({ head: "Production hall", tail: " A" });
    expect(splitName("Production hall B")).toEqual({ head: "Production hall", tail: " B" });
    expect(splitName("Boiler house")).toEqual({ head: "Boiler", tail: " house" });
    expect(splitName("Site")).toEqual({ head: "Site", tail: "" });
    expect(splitName("production-line-a")).toEqual({ head: "production-line", tail: "-a" });
  });

  it("never cuts a word, so a name with no short last word is left whole for the ellipsis", () => {
    expect(splitName("Reservoir")).toEqual({ head: "Reservoir", tail: "" });
    expect(splitName("Compressor room extension")).toEqual({ head: "Compressor room extension", tail: "" });
  });

  it("counts nodes in words", () => {
    expect(countWords(1)).toBe("1 node");
    expect(countWords(12)).toBe("12 nodes");
  });

  it("is one line at every zoom, and never grows past its cap in plant units however far the view is zoomed out", () => {
    expect(titleScale(1)).toBe(1);
    expect(titleScale(20)).toBe(capScale);
    expect(titleBandHeight(titleScale(1))).toBe(17);
    expect(titleBandHeight(titleScale(20))).toBeCloseTo(17 * capScale, 6);
  });

  it("reserves room for one header line at the text cap, whatever the area is called", () => {
    expect(headerReserve()).toBeCloseTo((17 + 4) * capScale, 6);
  });

  it("lifts the bounds of the plant to include the title above its first area", () => {
    const bounds = { x: 0, y: 100, width: 800, height: 400 };
    const lifted = boundsWithTitles(bounds, [100, 140], 2);
    expect(lifted.y).toBeCloseTo(100 - 21 * 2, 6);
    expect(lifted.y + lifted.height).toBe(500);
    expect(boundsWithTitles(bounds, [], 2)).toEqual(bounds);
  });
});
