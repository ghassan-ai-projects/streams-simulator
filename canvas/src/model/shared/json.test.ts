import { describe, expect, it } from "vitest";
import { isRecord, list, number, record, text, unescaped } from "./json";

describe("json narrowing", () => {
  it("accepts only plain objects as records", () => {
    expect(isRecord({ a: 1 })).toBe(true);
    expect([[], null, "x", 3, undefined].some(isRecord)).toBe(false);
    expect(record([1])).toEqual({});
    expect(record({ a: 1 })).toEqual({ a: 1 });
  });

  it("falls back to empty values for the wrong type", () => {
    expect(list([1])).toEqual([1]);
    expect(list("x")).toEqual([]);
    expect(text("a")).toBe("a");
    expect(text(1)).toBe("");
    expect(number(2.5)).toBe(2.5);
    expect([Number.NaN, Number.POSITIVE_INFINITY, "3", null].map(number)).toEqual([0, 0, 0, 0]);
  });
});

describe("pointer tokens", () => {
  it("decodes the escaped slash and tilde in that order", () => {
    expect(unescaped("a~1b~0c")).toBe("a/b~c");
    expect(unescaped("~01")).toBe("~1");
  });
});
