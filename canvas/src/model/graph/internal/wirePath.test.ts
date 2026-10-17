import { describe, expect, it } from "vitest";
import { labelPoint, wirePath } from "./wirePath";

describe("wire path", () => {
  it("rounds each corner with a quadratic curve and ends on the last point", () => {
    const path = wirePath([
      { x: 0, y: 0 },
      { x: 20, y: 0 },
      { x: 20, y: 40 },
      { x: 60, y: 40 },
    ]);
    expect(path).toBe("M 0 0 L 14 0 Q 20 0 20 6 L 20 34 Q 20 40 26 40 L 60 40");
  });

  it("shrinks a corner's radius to half of a short segment", () => {
    const path = wirePath(
      [
        { x: 0, y: 0 },
        { x: 4, y: 0 },
        { x: 4, y: 40 },
      ],
      6,
    );
    expect(path).toBe("M 0 0 L 2 0 Q 4 0 4 2 L 4 40");
    expect(wirePath([])).toBe("");
  });

  it("puts a label in the middle of the longest segment", () => {
    expect(
      labelPoint([
        { x: 0, y: 0 },
        { x: 10, y: 0 },
        { x: 10, y: 100 },
      ]),
    ).toEqual({ x: 10, y: 50 });
    expect(labelPoint([])).toEqual({ x: 0, y: 0 });
  });
});
