import { describe, expect, it } from "vitest";
import { directionOf, firstOf, nearest } from "./graphKeys";

const box = (x: number, y: number) => ({ x, y, width: 100, height: 60 });
const grid = new Map([
  ["a", box(0, 0)],
  ["b", box(300, 0)],
  ["c", box(0, 200)],
  ["d", box(300, 220)],
  ["far", box(1200, 10)],
]);

describe("graph keys", () => {
  it("maps arrow keys to directions and ignores the rest", () => {
    expect(["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "Enter"].map(directionOf)).toEqual([
      "left",
      "right",
      "up",
      "down",
      null,
    ]);
  });

  it("moves to the nearest node in the direction, preferring the straight line", () => {
    expect(nearest("a", "right", grid)).toBe("b");
    expect(nearest("a", "down", grid)).toBe("c");
    expect(nearest("d", "left", grid)).toBe("c");
    expect(nearest("d", "up", grid)).toBe("b");
    expect(nearest("b", "right", grid)).toBe("far");
  });

  it("stays put at an edge and for an unknown node", () => {
    expect(nearest("a", "left", grid)).toBeNull();
    expect(nearest("a", "up", grid)).toBeNull();
    expect(nearest("nobody", "up", grid)).toBeNull();
  });

  it("starts at the top-left node", () => {
    expect(firstOf(grid)).toBe("a");
    expect(
      firstOf(
        new Map([
          ["x", box(5, 5)],
          ["y", box(0, 5)],
        ]),
      ),
    ).toBe("y");
    expect(firstOf(new Map())).toBeNull();
  });
});
