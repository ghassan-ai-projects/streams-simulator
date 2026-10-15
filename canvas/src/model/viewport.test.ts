import { describe, expect, it } from "vitest";
import type { Viewport } from "./viewport";
import {
  steppedViewport,
  centreOf,
  followZoom,
  inView,
  openingFloor,
  openingViewport,
  sameViewport,
  textFloorPx,
  tierOf,
} from "./viewport";

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
    expect([1, 0.95, 0.6, 0.5, 0.45, 0.4, 0.39, 0.3, 0.2].map(tierOf)).toEqual([
      "full",
      "compact",
      "compact",
      "compact",
      "compact",
      "compact",
      "overview",
      "overview",
      "map",
    ]);
  });

  it("keeps readouts and port rows from the opening floor of 40 % up, one rule for every screen", () => {
    expect(tierOf(openingFloor)).toBe("compact");
    expect(textFloorPx).toBe(8);
  });

  it("finds a box's centre", () => {
    expect(centreOf({ x: 10, y: 20, width: 100, height: 40 })).toEqual({ x: 60, y: 40 });
  });

  it("opens on the whole plant centred when it fits, and on its top-left corner at the minimum zoom when not", () => {
    const limits = { pad: 10, headroom: 20, minZoom: 0.5, maxZoom: 1 };
    expect(openingViewport({ x: 0, y: 0, width: 390, height: 290 }, screen, limits)).toEqual({
      x: 205,
      y: 165,
      zoom: 1,
    });
    expect(openingViewport({ x: 100, y: 50, width: 1560, height: 580 }, screen, limits)).toEqual({
      x: -40,
      y: 140,
      zoom: 0.5,
    });
    expect(openingViewport({ x: 100, y: 50, width: 3000, height: 2000 }, screen, limits)).toEqual({
      x: -40,
      y: 5,
      zoom: 0.5,
    });
  });

  it("opens on the whole plant, nothing cut, when the floor is low enough to fit it", () => {
    const limits = { pad: 16, headroom: 24, minZoom: 0.25, maxZoom: 1 };
    const bounds = { x: 20, y: 30, width: 3000, height: 2000 };
    const view = openingViewport(bounds, screen, limits);
    expect(view.zoom).toBeGreaterThanOrEqual(0.25);
    expect(view.zoom).toBeLessThan(0.5);
    expect(bounds.x * view.zoom + view.x).toBeGreaterThanOrEqual(16);
    expect((bounds.x + bounds.width) * view.zoom + view.x).toBeLessThanOrEqual(screen.width - 16 + 1e-6);
    expect(bounds.y * view.zoom + view.y).toBeGreaterThanOrEqual(40);
    expect((bounds.y + bounds.height) * view.zoom + view.y).toBeLessThanOrEqual(screen.height - 16 + 1e-6);
  });

  it("opens a plant too large for the floor on its anchor, fully on screen, never past the plant's edges", () => {
    const limits = { pad: 10, headroom: 20, minZoom: 0.5, maxZoom: 1 };
    const bounds = { x: 0, y: 0, width: 3000, height: 2000 };
    const onScreen = (anchor: { x: number; y: number; width: number; height: number }, view: Viewport) =>
      anchor.x * view.zoom + view.x >= 0 &&
      anchor.y * view.zoom + view.y >= 0 &&
      (anchor.x + anchor.width) * view.zoom + view.x <= screen.width &&
      (anchor.y + anchor.height) * view.zoom + view.y <= screen.height;
    for (const anchor of [
      { x: 1400, y: 900, width: 200, height: 100 },
      { x: 2800, y: 1900, width: 200, height: 100 },
      { x: 0, y: 0, width: 200, height: 100 },
    ]) {
      const view = openingViewport(bounds, screen, limits, anchor);
      expect(view.zoom).toBe(0.5);
      expect(onScreen(anchor, view)).toBe(true);
      expect(view.x).toBeLessThanOrEqual(10);
      expect(view.x + 3000 * 0.5).toBeGreaterThanOrEqual(screen.width - 10 - 1e-6);
    }
    const middle = openingViewport(bounds, screen, limits, { x: 1400, y: 900, width: 200, height: 100 });
    expect(middle.x + 1500 * 0.5).toBeCloseTo(400);
    expect(middle.y + 950 * 0.5).toBeCloseTo(310);
  });

  it("treats viewports within half a pixel as the same", () => {
    expect(sameViewport({ x: 1, y: 2, zoom: 0.5 }, { x: 1.2, y: 2.3, zoom: 0.5 })).toBe(true);
    expect(sameViewport({ x: 1, y: 2, zoom: 0.5 }, { x: 1, y: 2, zoom: 0.6 })).toBe(false);
  });

  it("zooms about the centre of the canvas and keeps the content's top-left titles in view", () => {
    const screen = { width: 800, height: 600 };
    const content = { x: 0, y: 0, width: 2000, height: 1000 };
    const start = { x: 16, y: 16, zoom: 0.4 };
    const next = steppedViewport(start, screen, 0.48, content, 16);
    expect(next).toEqual({ x: 16, y: 16, zoom: 0.48 });
    const panned = steppedViewport({ x: -400, y: 16, zoom: 1 }, screen, 1.2, content, 16);
    expect(panned.x).toBe(400 - (400 + 400) * 1.2);
  });

  it("pulls small content back inside the canvas when zooming leaves it hanging off the right", () => {
    const screen = { width: 800, height: 600 };
    const content = { x: 0, y: 0, width: 300, height: 200 };
    const next = steppedViewport({ x: 480, y: 16, zoom: 1 }, screen, 1.2, content, 16);
    expect(next.x).toBe(800 - 16 - 360);
  });
});
