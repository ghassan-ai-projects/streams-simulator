import { describe, expect, it } from "vitest";
import { markNoteLines, noteLineCount, widestNotes } from "./noteLines";
import type { PlanMark } from "./types";

const mark = (fault: string, effector: string): PlanMark => ({
  status: fault === "" ? null : "fault",
  count: 0,
  fault,
  effector,
});

describe("the lines a node reserves for what is said about it", () => {
  it("is one for nothing, a short cause or a short outcome, and adds the wrapped lines of each", () => {
    expect(noteLineCount("", "")).toBe(1);
    expect(noteLineCount("scaling", "")).toBe(1);
    expect(noteLineCount("", "stop · accepted")).toBe(1);
    expect(noteLineCount("fan failure", "set fan speed · accepted")).toBe(4);
  });

  it("is counted per node from the fault and the command a run shows on it, and none for a node it says nothing of", () => {
    const lines = markNoteLines(
      new Map([
        ["CT-1", mark("fan failure", "set fan speed · accepted")],
        ["P-1", mark("", "")],
        ["CH-1", mark("", "set setpoint · accepted")],
      ]),
    );
    expect(lines.get("CT-1")).toBe(4);
    expect(lines.get("CH-1")).toBe(2);
    expect(lines.has("P-1")).toBe(false);
    expect(markNoteLines(undefined).size).toBe(0);
  });

  it("keeps, per node, the larger of two counts", () => {
    const wide = widestNotes(
      new Map([
        ["A", 4],
        ["B", 1],
      ]),
      new Map([
        ["A", 2],
        ["B", 3],
        ["C", 2],
      ]),
    );
    expect([...wide]).toEqual([
      ["A", 4],
      ["B", 3],
      ["C", 2],
    ]);
    expect(widestNotes(undefined, new Map([["A", 2]])).get("A")).toBe(2);
  });
});
