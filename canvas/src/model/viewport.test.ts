import { describe, expect, it } from "vitest";
import type { Tier } from "./viewport";
import { anchorsOnly, centreOf, followZoom, inView, openingViewport, sameViewport, tierOf } from "./viewport";

describe("viewport", () => {
  const screen = { width: 800, height: 600 };
  it("moves the view for a selected box only when it is off screen, and then to at least the focus zoom", () => {
    const box = { x: 10, y: 10, width: 100, height: 50 };
    expect(followZoom(box, { x: 0, y: 0, zoom: 0.62 }, screen, 1)).toBeNull();
    expect(followZoom(box, { x: -200, y: 0, zoom: 0.62 }, screen, 1)).toBe(1);
    expect(followZoom(box, { x: -200, y: 0, zoom: 2 }, screen, 1)).toBe(2);
    expect(followZoom(box, { x: -200, y: 0, zoom: 0.62 }, screen, 0)).toBe(0.62);
  });

  it("knows whether a box is fully on screen at the current pan and zoom", () => {
    expect(inView({ x: 10, y: 10, width: 100, height: 50 }, { x: 0, y: 0, zoom: 1 }, screen)).toBe(true);
    expect(inView({ x: 10, y: 10, width: 100, height: 50 }, { x: -20, y: 0, zoom: 1 }, screen)).toBe(false);
    expect(inView({ x: 700, y: 10, width: 200, height: 50 }, { x: 0, y: 0, zoom: 0.5 }, screen)).toBe(true);
    expect(inView({ x: 10, y: 590, width: 10, height: 20 }, { x: 0, y: 0, zoom: 1 }, screen)).toBe(false);
  });

  it("picks the node tier for a zoom, so text never shrinks below its size", () => {
    expect([1, 0.95, 0.6, 0.5, 0.3, 0.2].map(tierOf)).toEqual([
      "full",
      "compact",
      "compact",
      "compact",
      "overview",
      "map",
    ]);
    expect(["full", "compact", "overview", "map"].map((tier) => anchorsOnly(tier as Tier))).toEqual([
      false,
      true,
      true,
      true,
    ]);
  });

  it("finds a box's centre", () => {
    expect(centreOf({ x: 10, y: 20, width: 100, height: 40 })).toEqual({ x: 60, y: 40 });
  });

  it("opens on the whole plant at the top when it fits, and on its top-left corner at the minimum zoom when not", () => {
    const limits = { pad: 10, headroom: 20, minZoom: 0.5, maxZoom: 1 };
    expect(openingViewport({ x: 0, y: 0, width: 390, height: 290 }, screen, limits)).toEqual({
      x: 205,
      y: 30,
      zoom: 1,
    });
    expect(openingViewport({ x: 100, y: 50, width: 1560, height: 580 }, screen, limits)).toEqual({
      x: -40,
      y: 5,
      zoom: 0.5,
    });
    expect(openingViewport({ x: 100, y: 50, width: 3000, height: 2000 }, screen, limits)).toEqual({
      x: -40,
      y: 5,
      zoom: 0.5,
    });
  });

  it("treats viewports within half a pixel as the same", () => {
    expect(sameViewport({ x: 1, y: 2, zoom: 0.5 }, { x: 1.2, y: 2.3, zoom: 0.5 })).toBe(true);
    expect(sameViewport({ x: 1, y: 2, zoom: 0.5 }, { x: 1, y: 2, zoom: 0.6 })).toBe(false);
  });
});
