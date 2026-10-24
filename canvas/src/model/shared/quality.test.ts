import { describe, expect, it } from "vitest";
import { qualityWords } from "./quality";

describe("quality words", () => {
  it.each([
    ["", ""],
    ["suspect", "suspect"],
    ["out_of_range", "out of range"],
    ["stale", "stale"],
    ["disconnected", "disconnected"],
  ] as const)("words %j as %j, and good as nothing at all", (quality, expected) => {
    expect(qualityWords(quality)).toBe(expected);
  });
});
