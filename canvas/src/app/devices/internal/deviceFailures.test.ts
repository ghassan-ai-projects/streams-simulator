import { describe, expect, it } from "vitest";
import { blankKind } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";
import { refusalWords } from "./deviceFailures";

const diagnostic = { message: "refused", displayCode: "kind_failed" };

describe("the words of a refusal", () => {
  const kind = blankKind("fan");

  it("puts each issue in plain words", () => {
    const failure: DeviceFailure = {
      kind: "refused",
      diagnostic,
      issues: [
        { pointer: "/kind", message: "raw", keyword: "pattern" },
        { pointer: "/metrics", message: "raw", keyword: "twin_member_required", member: "metrics" },
      ],
    };
    expect(refusalWords(kind, failure)).toBe("Use letters, digits and . _ : - only Needs at least one reading");
  });

  it("falls back on the binary's message when there is no issue to word", () => {
    expect(refusalWords(kind, { kind: "refused", diagnostic, issues: [] })).toBe("refused");
    expect(refusalWords(kind, { kind: "unavailable", diagnostic })).toBe("refused");
  });
});
