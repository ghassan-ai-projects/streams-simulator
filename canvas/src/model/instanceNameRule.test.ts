import { describe, expect, it } from "vitest";
import { instanceNameRule } from "./instanceNameRule";

describe("instance name rule", () => {
  it("keeps the list of allowed punctuation on one line", () => {
    const text = instanceNameRule();
    expect(text).toContain("and . _ : / -,");
    expect(text).toMatch(/^An instance name/);
  });
});
