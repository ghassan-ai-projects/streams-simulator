import { describe, expect, it } from "vitest";
import { failureLine } from "./failureLine";

describe("failure line", () => {
  it("puts the display code after the diagnostic message", () => {
    expect(failureLine({ message: "The bridge did not answer.", displayCode: "unreachable" })).toBe(
      "The bridge did not answer. (unreachable)",
    );
  });
});
