import { describe, expect, it } from "vitest";
import { sameValue } from "./sameValue";

describe("same value", () => {
  it.each([
    [1, 1, true],
    [1, 2, false],
    ["a", "a", true],
    [null, null, true],
    [null, {}, false],
    [Number.NaN, Number.NaN, true],
    [[1, 2], [1, 2], true],
    [[1, 2], [1, 2, 3], false],
    [[1, 2], { 0: 1, 1: 2 }, false],
    [{ a: 1, b: { c: [1] } }, { a: 1, b: { c: [1] } }, true],
    [{ a: 1 }, { a: 1, b: undefined }, false],
    [{ a: undefined }, { b: undefined }, false],
    [new Set(["a", "b"]), new Set(["b", "a"]), true],
    [new Set(["a"]), new Set(["a", "b"]), false],
    [new Set(["a"]), ["a"], false],
    [new Map([["a", 1]]), new Map([["a", 1]]), false],
  ])("compares %j with %j as %s", (a, b, same) => {
    expect(sameValue(a, b)).toBe(same);
  });

  it("is quick about the very same object", () => {
    const big = { rows: Array.from({ length: 1000 }, (_, at) => ({ at })) };
    expect(sameValue(big, big)).toBe(true);
  });
});
