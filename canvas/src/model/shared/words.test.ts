import { describe, expect, it } from "vitest";
import { failureLine, humanise, spaced } from "./words";

describe("words", () => {
  it("puts the display code after the diagnostic message", () => {
    expect(failureLine({ message: "The bridge did not answer.", displayCode: "unreachable" })).toBe(
      "The bridge did not answer. (unreachable)",
    );
  });

  it("spaces underscores and humanises ids", () => {
    expect(spaced("flow_rate_high")).toBe("flow rate high");
    expect(humanise("plant__flow_rate")).toBe("flow rate");
  });
});
