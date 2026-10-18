import { describe, expect, it } from "vitest";
import {
  birthWords,
  capturePath,
  captureSizeProblem,
  captureTextProblem,
  claimWords,
  comparisonWords,
  fileSafeName,
  originWords,
  refusalLines,
} from "./captureWords";

type Comparison = Parameters<typeof comparisonWords>[0];

describe("the words of a captured announcement", () => {
  it("says where it came from, with a word and not a flag", () => {
    expect(originWords("hardware")).toBe("Real board");
    expect(originWords("simulated")).toBe("Simulated");
  });

  it("numbers the announcement in plain words, and says it is a claim", () => {
    expect(birthWords(0)).toBe("Announcement number 0");
    expect(birthWords(12)).toBe("Announcement number 12");
    expect(claimWords).toBe("A board's announcement is a claim. Nothing is saved until you review and accept it.");
  });

  const comparisons: [Comparison, string][] = [
    [{ status: "new" }, "New device"],
    [{ status: "same", file: "kinds/pump.kind.json" }, "Same as pump.kind.json"],
    [{ status: "differs", file: "/work/kinds/pump.kind.json" }, "Differs from pump.kind.json"],
    [{ status: "same", file: "pump.kind.json" }, "Same as pump.kind.json"],
  ];

  it.each(comparisons)("words the comparison %j as %j", (comparison, words) => {
    expect(comparisonWords(comparison)).toBe(words);
  });
});

describe("what may be pasted", () => {
  it("accepts up to 1 MiB and refuses more, in a sentence that says what to do instead", () => {
    expect(captureSizeProblem(1024 * 1024)).toBe("");
    expect(captureSizeProblem(1024 * 1024 + 1)).toMatch(/larger than 1 MiB.*recorded run/);
    expect(captureTextProblem("{}")).toBe("");
    expect(captureTextProblem("x".repeat(1024 * 1024 + 1))).not.toBe("");
  });

  it("counts bytes, not letters", () => {
    expect(captureTextProblem("é".repeat(600 * 1024))).not.toBe("");
    expect(captureTextProblem("e".repeat(600 * 1024))).toBe("");
  });
});

describe("where a captured kind is saved", () => {
  it("uses the file of the kind that is already saved, or the device's own path", () => {
    const kinds = [{ path: "kinds/old-pump.kind.json", kind: "pump" }];
    expect(capturePath(kinds, "pump")).toBe("kinds/old-pump.kind.json");
    expect(capturePath(kinds, "fan")).toBe("kinds/fan.kind.json");
    expect(capturePath([], "fan")).toBe("kinds/fan.kind.json");
  });

  it.each([
    ["pump", true],
    ["thermal_sensor", true],
    ["", false],
    ["a/b", false],
    ["a\\b", false],
    ["..", false],
    [".hidden", false],
  ])("takes the name %j as a file name: %j", (name, safe) => {
    expect(fileSafeName(name)).toBe(safe);
  });
});

describe("a refused announcement", () => {
  it("words each issue and keeps the binary's message beside it", () => {
    const issues = [
      { pointer: "/kinds/0/kind_digest", message: "does not hold", keyword: "kind_digest_mismatch" },
      { pointer: "", message: "none here", keyword: "no_birth" },
    ];
    expect(refusalLines(issues)).toEqual([
      {
        text: "The announcement does not match its own fingerprint, so it cannot be trusted",
        message: "does not hold",
      },
      { text: "No announcement from a board was found", message: "none here" },
    ]);
  });
});
