import { describe, expect, it } from "vitest";
import { placeChips, timeTicks } from "./timeLayout";

describe("event lane layout", () => {
  it("centres each chip on its event and puts close events on separate rows, reusing a row once free", () => {
    const lane = { width: 100, rows: 2, gap: 4 };
    const rows = (positions: number[]) =>
      placeChips(
        positions,
        positions.map(() => 20),
        lane,
      ).map((chip) => chip.row);
    expect(rows([10, 40, 41])).toEqual([0, 0, 1]);
    expect(placeChips([50], [20], lane)).toEqual([{ left: 40, row: 0 }]);
    expect(rows([0, 1, 2, 3])).toEqual([0, 1, -1, -1]);
    expect(rows([])).toEqual([]);
    expect(placeChips([95], [40], lane)).toEqual([{ left: 60, row: 0 }]);
  });

  it("puts time ticks on round steps, at most nine across the run", () => {
    expect(timeTicks(0, 8 * 3600e9)).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8].map((h) => h * 3600e9));
    expect(timeTicks(0, 1800e9)).toEqual([0, 600e9, 1200e9, 1800e9]);
    expect(timeTicks(100e9, 3700e9).at(0)).toBe(600e9);
    expect(timeTicks(0, 0)).toEqual([]);
  });
});
