import { describe, expect, it } from "vitest";
import type { Tier } from "./viewport";
import { anchorsOnly, centreOf, inView, tierOf } from "./viewport";

describe("viewport", () => {
  const screen = { width: 800, height: 600 };
  it("knows whether a box is fully on screen at the current pan and zoom", () => {
    expect(inView({ x: 10, y: 10, width: 100, height: 50 }, { x: 0, y: 0, zoom: 1 }, screen)).toBe(true);
    expect(inView({ x: 10, y: 10, width: 100, height: 50 }, { x: -20, y: 0, zoom: 1 }, screen)).toBe(false);
    expect(inView({ x: 700, y: 10, width: 200, height: 50 }, { x: 0, y: 0, zoom: 0.5 }, screen)).toBe(true);
    expect(inView({ x: 10, y: 590, width: 10, height: 20 }, { x: 0, y: 0, zoom: 1 }, screen)).toBe(false);
  });

  it("picks the node tier for a zoom, so text never shrinks below its size", () => {
    expect([1, 0.85, 0.6, 0.5, 0.3, 0.2].map(tierOf)).toEqual([
      "full",
      "full",
      "compact",
      "compact",
      "overview",
      "map",
    ]);
    expect(["full", "compact", "overview", "map"].map((tier) => anchorsOnly(tier as Tier))).toEqual([
      false,
      false,
      true,
      true,
    ]);
  });

  it("finds a box's centre", () => {
    expect(centreOf({ x: 10, y: 20, width: 100, height: 40 })).toEqual({ x: 60, y: 40 });
  });
});
