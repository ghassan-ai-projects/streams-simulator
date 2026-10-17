import { describe, expect, it } from "vitest";
import { groupBy, tally } from "./collect";

describe("collect", () => {
  it("groups items by key in first-seen key order, keeping item order", () => {
    const groups = groupBy(["b1", "a1", "b2", "a2"], (item) => item[0]);
    expect([...groups.keys()]).toEqual(["b", "a"]);
    expect(groups.get("b")).toEqual(["b1", "b2"]);
    expect(groups.get("a")).toEqual(["a1", "a2"]);
  });

  it("tallies items by key", () => {
    const counts = tally(["x", "y", "x", "x"], (item) => item);
    expect([...counts]).toEqual([
      ["x", 3],
      ["y", 1],
    ]);
  });

  it("returns empty maps for no items", () => {
    expect(groupBy([], (item) => item).size).toBe(0);
    expect(tally([], (item) => item).size).toBe(0);
  });
});
