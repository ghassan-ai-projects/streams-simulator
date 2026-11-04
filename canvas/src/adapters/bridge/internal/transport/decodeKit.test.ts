import { describe, expect, it } from "vitest";
import { finiteOrNull } from "./decodeKit";

describe("finiteOrNull", () => {
  it("keeps finite numbers", () => {
    expect([finiteOrNull(0), finiteOrNull(-2.5)]).toEqual([0, -2.5]);
  });

  it("rejects everything else", () => {
    const rejected = [NaN, Infinity, -Infinity, "1", null, undefined, {}, true].map(finiteOrNull);
    expect(rejected.every((value) => value === null)).toBe(true);
  });
});
