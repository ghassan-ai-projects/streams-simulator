import { describe, expect, it } from "vitest";
import { alarmRow, modeWords } from "./words";

describe("review words", () => {
  it("names a known alarm figure and explains it", () => {
    expect(alarmRow("flood_intervals", "2")).toEqual([
      "Overload spells",
      "2",
      "10-minute spells where alarms came too fast for an operator to follow.",
    ]);
  });

  it("spaces an unknown alarm key and gives it no explanation", () => {
    expect(alarmRow("shelved_total", "4")).toEqual(["shelved total", "4", ""]);
  });

  it("puts the Evaluate rule after the mode's name", () => {
    expect(modeWords.evaluate).toBe(`Evaluate: ${modeWords.evaluateRule}`);
    expect(modeWords.explore).toBe("Explore: the truth is visible; not for scoring.");
  });
});
