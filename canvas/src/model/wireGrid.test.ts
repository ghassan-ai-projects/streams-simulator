import { describe, expect, it } from "vitest";
import { buildGrid } from "./wireGrid";

const ones = (bits: Uint8Array): number[] => [...bits].flatMap((bit, edge) => (bit === 1 ? [edge] : []));

describe("wire grid", () => {
  it("marks the grid lines that run alongside a frame inside its margin band, and no others", () => {
    const grid = buildGrid([], [180, 200, 212, 260], [0, 50, 100, 150], {
      frames: [{ x: 0, y: 0, width: 200, height: 100 }],
      band: 12,
    });
    expect(ones(grid.hugV)).toEqual([4, 5, 6]);
    expect(ones(grid.hugH)).toEqual([0, 2, 4, 6]);
  });

  it("leaves a line one band away and the lines outside the frame's span alone", () => {
    const grid = buildGrid([], [212, 260], [50, 150], {
      frames: [{ x: 0, y: 0, width: 200, height: 100 }],
      band: 12,
    });
    expect(ones(grid.hugV)).toEqual([]);
    expect(ones(grid.hugH)).toEqual([]);
  });

  it("marks nothing when no frames are given", () => {
    const grid = buildGrid([], [0, 100], [0, 100]);
    expect(ones(grid.hugH)).toEqual([]);
    expect(ones(grid.hugV)).toEqual([]);
  });
});
