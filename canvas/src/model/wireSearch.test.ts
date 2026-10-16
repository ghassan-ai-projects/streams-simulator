import { describe, expect, it } from "vitest";
import { buildGrid, nearestIndex } from "./wireGrid";
import { emptyOccupancy, findPath, occupy } from "./wireSearch";

describe("wire occupancy", () => {
  it("closes the lanes within one pitch of a used lane, so the next wire keeps its distance", () => {
    const grid = buildGrid([], [0, 100, 200], [0, 100]);
    const occupancy = emptyOccupancy(grid);
    const row = (iy: number): readonly number[] => [(0 * grid.ny + iy) * 2, (2 * grid.ny + iy) * 2];
    occupy(grid, occupancy, row(0), 150);
    expect(occupancy.usedH[0]).toBe(1);
    expect(occupancy.usedH[1]).toBe(1);
    expect(occupancy.usedH[3]).toBe(1);
    expect(occupancy.usedH[5]).toBe(1);
  });

  it("leaves the neighbouring lanes open when no pitch is asked for", () => {
    const grid = buildGrid([], [0, 100, 200], [0, 100]);
    const occupancy = emptyOccupancy(grid);
    occupy(grid, occupancy, [0, 2 * grid.ny * 2]);
    expect(occupancy.usedH[0]).toBe(1);
    expect(occupancy.usedH[1]).toBe(0);
  });
});

describe("wire search", () => {
  it("takes the lane outside a frame's margin when the hugged lane beside it would cost more", () => {
    const frame = { x: 0, y: 0, width: 200, height: 100 };
    const xs = [206, 224];
    const ys = [-20, 50, 120];
    const grid = buildGrid([], xs, ys, { frames: [frame], band: 12 });
    const ends = {
      start: nearestIndex(xs, 224) * grid.ny + nearestIndex(ys, -20),
      goal: nearestIndex(xs, 206) * grid.ny + nearestIndex(ys, 120),
    };
    const path = findPath(grid, emptyOccupancy(grid), ends, { start: true, goal: false });
    expect(path).not.toBeNull();
    const columnOf = (vertex: number): number => Math.floor(vertex / grid.ny);
    const huggedColumn = nearestIndex(xs, 206);
    const vertices = path ?? [];
    const approached = vertices.slice(0, -1);
    expect(approached.every((vertex) => columnOf(vertex) !== huggedColumn)).toBe(true);
    expect(columnOf(vertices.at(-1) ?? -1)).toBe(huggedColumn);
  });
});
