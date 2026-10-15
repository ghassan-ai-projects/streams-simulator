import { describe, expect, it } from "vitest";
import {
  boundsWithTitles,
  capScale,
  countWords,
  headerReserve,
  titleBandHeight,
  titleLines,
  titleScale,
} from "./frameTitle";

describe("frame title", () => {
  it("counts nodes in words", () => {
    expect(countWords(1)).toBe("1 node");
    expect(countWords(12)).toBe("12 nodes");
  });

  it("grows by one line for each time the name and count together overflow the frame", () => {
    expect(titleLines("Hall", 2, 300, 1)).toBe(1);
    expect(titleLines("Production hall A", 1, 280, 1)).toBe(1);
    expect(titleLines("Production hall A", 1, 280, capScale)).toBe(2);
    expect(titleLines("Production hall A", 1, 220, capScale)).toBe(3);
    expect(titleLines("A very long area name that keeps going and going", 3, 280, 1)).toBeGreaterThanOrEqual(2);
  });

  it("never lets a title's text grow past its cap in plant units, however far the view is zoomed out", () => {
    expect(titleScale(1)).toBe(1);
    expect(titleScale(20)).toBe(capScale);
    expect(titleBandHeight("Hall", 1, 300, titleScale(20))).toBeCloseTo(17 * capScale, 6);
  });

  it("reserves room for the tallest header the layout can see at the text cap", () => {
    const short = headerReserve("Hall", 1, 400);
    const wrapped = headerReserve("Production hall A", 1, 280);
    expect(wrapped).toBeGreaterThan(short);
    expect(short).toBeGreaterThan(0);
  });

  it("lifts the bounds of the plant to include the tallest title above its first area", () => {
    const bounds = { x: 0, y: 100, width: 800, height: 400 };
    const titles = [
      { y: 100, width: 300, title: "Hall", count: 1 },
      { y: 100, width: 300, title: "Production hall A", count: 1 },
    ];
    const lifted = boundsWithTitles(bounds, titles, 2);
    expect(lifted.y).toBeLessThan(100 - 34);
    expect(lifted.y + lifted.height).toBe(500);
    expect(boundsWithTitles(bounds, [], 2)).toEqual(bounds);
  });
});
