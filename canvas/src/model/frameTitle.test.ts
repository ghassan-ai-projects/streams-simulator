import { describe, expect, it } from "vitest";
import { boundsWithTitles, capScale, countWords, headerReserve, titleBandHeight, titleScale } from "./frameTitle";

describe("frame title", () => {
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
