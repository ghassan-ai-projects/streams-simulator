import { describe, expect, it } from "vitest";
import {
  birthWords,
  captureSizeProblem,
  captureTextProblem,
  claimedWords,
  claimWords,
  collisionWords,
  comparisonWords,
  originWords,
  refusalLines,
  refusalOutline,
  scanNotices,
  sourceWords,
  unreadableWords,
} from "./captureWords";

type Comparison = Parameters<typeof comparisonWords>[0];

describe("the words of a captured announcement", () => {
  it("says where it came from, with a word and not a flag", () => {
    expect(originWords("hardware")).toBe("Real board");
    expect(originWords("simulated")).toBe("Simulated");
  });

  it("counts the board's restarts in plain words, and says nothing for the first announcement", () => {
    expect(birthWords(0)).toBe("");
    expect(birthWords(12)).toBe("Board restart count: 12");
    expect(claimWords).toBe("A board's announcement is a claim. Nothing is saved until you review and accept it.");
  });

  it("says what was read", () => {
    expect(sourceWords({ content: "{}" })).toBe("the pasted text");
    expect(sourceWords({ run: "w-1" })).toBe("the run w-1");
  });

  const comparisons: [Comparison, string][] = [
    [{ status: "new" }, "New device"],
    [{ status: "same", file: "kinds/pump.kind.json" }, "Same as pump.kind.json"],
    [{ status: "differs", file: "/work/kinds/pump.kind.json" }, "Differs from pump.kind.json"],
    [{ status: "same", file: "pump.kind.json" }, "Same as pump.kind.json"],
    [{ status: "unreadable", file: "kinds/pump.kind.json" }, "Cannot compare"],
  ];

  it.each(comparisons)("words the comparison %j as %j", (comparison, words) => {
    expect(comparisonWords(comparison)).toBe(words);
  });

  it("says a name that differs only by capital letters, and a saved file that cannot be read", () => {
    expect(collisionWords("kinds/Pump.kind.json")).toBe(
      "Looks like Pump.kind.json, which differs only by capital letters.",
    );
    expect(unreadableWords("kinds/pump.kind.json")).toMatch(/^The saved file pump\.kind\.json cannot be read/);
  });
});

describe("what may be pasted", () => {
  it("accepts up to 1 MiB and refuses more, in a sentence that says what to do instead", () => {
    expect(captureSizeProblem(1024 * 1024)).toBe("");
    expect(captureSizeProblem(1024 * 1024 + 1)).toMatch(/larger than 1 MiB.*recorded run/);
    expect(captureTextProblem("{}")).toBe("");
    expect(captureTextProblem("x".repeat(1024 * 1024 + 1))).not.toBe("");
  });

  it("counts bytes, not letters, and does not encode what cannot be too large", () => {
    expect(captureTextProblem("é".repeat(600 * 1024))).not.toBe("");
    expect(captureTextProblem("e".repeat(600 * 1024))).toBe("");
    expect(captureTextProblem("é".repeat(349525))).toBe("");
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

  it.each([
    ["no_birth", "No announcement found", /no announcement from a board/],
    ["json", "That is not valid JSON", /cannot be read as JSON/],
    ["kind_digest_mismatch", "The board's announcement was not accepted", /own fingerprint/],
    ["sealed_stream", "The board's announcement was not accepted", /sealed part of a run/],
    ["stream_not_observer", "The board's announcement was not accepted", /observer's side/],
    ["something_new", "The board's announcement was not accepted", /cannot be trusted/],
  ])("leads with plain words for %s", (keyword, heading, sentence) => {
    const outline = refusalOutline("refused", [keyword]);
    expect(outline.heading).toBe(heading);
    expect(outline.sentence).toMatch(sentence);
    expect(outline.next).toMatch(/announcement|run/);
  });

  it("takes the first issue it has words for", () => {
    expect(refusalOutline("refused", ["unknown", "no_birth", "json"]).heading).toBe("No announcement found");
    expect(refusalOutline("refused", []).heading).toBe("The board's announcement was not accepted");
  });

  it.each([
    ["missing", "No announcement found"],
    ["unavailable", "The bridge is unavailable"],
    ["unreadable", "The announcement could not be read"],
    ["conflict", "The board's announcement was not accepted"],
  ])("words a failure of kind %s as %j whatever the issues", (kind, heading) => {
    expect(refusalOutline(kind, ["no_birth"]).heading).toBe(heading);
  });
});

describe("what a log holds beside the announcement taken", () => {
  const scan = (over: object = {}) => ({ birthsSeen: 1, taken: { line: 1 }, laterBirths: [], issues: [], ...over });

  it("says nothing about a single clean announcement", () => {
    expect(scanNotices(scan())).toEqual([]);
  });

  it("says which of several announcements was taken, and whether the later ones differ", () => {
    const later = [{ sameContent: true }];
    expect(scanNotices(scan({ birthsSeen: 2, taken: { line: 3 }, laterBirths: later }))).toEqual([
      "This log holds 2 announcements; showing the first (line 3). Later ones describe the same devices.",
    ]);
    const differ = [{ sameContent: true }, { sameContent: false }];
    expect(scanNotices(scan({ birthsSeen: 3, laterBirths: differ }))[0]).toMatch(
      /holds 3 announcements.*\(line 1\)\. Later ones differ\./,
    );
  });

  it("says what the scan found beside the kinds, once each", () => {
    const issues = [
      { message: "m", keyword: "birth_line_unreadable" },
      { message: "m", keyword: "birth_line_unreadable" },
      { message: "m", keyword: "capability_digest_mismatch" },
      { message: "m", keyword: "scan_truncated" },
      { message: "m" },
    ];
    const notices = scanNotices(scan({ birthsSeen: 2, laterBirths: [{ sameContent: true }], issues }));
    expect(notices).toEqual([
      "This log holds at least 2 announcements; showing the first (line 1). Later ones describe the same devices.",
      "A line that names an announcement could not be read.",
      "The board's capability fingerprint does not match its devices; treat this announcement with care.",
      "Only the start of this log was read.",
    ]);
  });

  it("still says a rebirth differs when no later birth was listed", () => {
    const issues = [{ message: "m", keyword: "rebirth_differs" }];
    expect(scanNotices(scan({ issues }))).toEqual(["A later announcement describes other devices than the first."]);
  });
});

describe("what the board claimed and nothing verified", () => {
  it("names the claimed members in plain words, once each", () => {
    const claims = [
      "/commands/0/risk_class",
      "/commands/0/safe_stop",
      "/commands/1/risk_class",
      "/commands/0/args_schema/properties/pct/minimum",
      "/commands/0/args_schema/properties/pct/maximum",
    ];
    expect(claimedWords(claims)).toBe(
      "Taken from the board's announcement, not verified: risk class, safe stop, argument limits.",
    );
  });

  it("says nothing when nothing was claimed", () => {
    expect(claimedWords([])).toBe("");
    expect(claimedWords(["", "/"])).toBe("");
  });
});
