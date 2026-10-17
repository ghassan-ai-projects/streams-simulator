import { describe, expect, it } from "vitest";
import { clockOf, durationOf } from "./duration";

describe("simulated clock", () => {
  it("prints simulated seconds as hours, minutes and seconds", () => {
    expect(clockOf(0)).toBe("00:00:00");
    expect(clockOf(3725)).toBe("01:02:05");
    expect(clockOf(-3)).toBe("00:00:00");
  });
});

describe("durations in words", () => {
  it("says how long in words", () => {
    expect([durationOf(0), durationOf(45), durationOf(1800), durationOf(5400), durationOf(86400)]).toEqual([
      "",
      "45 s",
      "30 min",
      "1 h 30 min",
      "24 h",
    ]);
  });
});
