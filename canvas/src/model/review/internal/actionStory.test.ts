import { describe, expect, it } from "vitest";
import { report } from "../../../test/runFixture";
import { storyOf, storyWindow } from "./actionStory";

describe("the story of one effector call", () => {
  it("reads the exact request and what the binary reported for it", () => {
    const story = storyOf(report, { atNs: 1_200e9, entity: "CH-1" });
    expect(story).toEqual({
      entity: "CH-1",
      effector: "start",
      requestedAt: "00:20:00",
      args: [
        { name: "chiller_id", value: "CH-1" },
        { name: "level", value: "0.5" },
      ],
      outcome: "applied · 300 ms",
      applied: true,
      source: null,
      atNs: 1_200e9,
    });
  });

  it("names a refusal with its reason and lists no arguments the call did not carry", () => {
    const story = storyOf(report, { atNs: 1_400e9, entity: "CH-1" });
    expect(story).toMatchObject({ outcome: "accepted, no effect · silent_no_effect", args: [], applied: false });
  });

  it("says a call the operator gave during the run was given by the operator", () => {
    const given = { ...report, effectorCalls: [{ ...report.effectorCalls[0], immediate: true }] } as typeof report;
    expect(storyOf(given, { atNs: 1_200e9, entity: "CH-1" })?.source).toBe("by the operator during the run");
  });

  it("has no story when nothing is picked or the pick is not an effector call", () => {
    expect(storyOf(report, null)).toBeNull();
    expect(storyOf(report, { atNs: 600e9, entity: "CT-1" })).toBeNull();
  });
});

describe("the minute around a call", () => {
  it("moves one minute before or after, kept inside the run", () => {
    expect(storyWindow(report, 1_200e9, "before")).toBe(1_140e9);
    expect(storyWindow(report, 1_200e9, "after")).toBe(1_260e9);
    expect(storyWindow(report, 30e9, "before")).toBe(0);
    expect(storyWindow(report, 3_590e9, "after")).toBe(3_600e9);
  });

  it("does not clamp to an end it does not know", () => {
    expect(storyWindow({ ...report, endNs: 0 }, 100e9, "after")).toBe(160e9);
  });
});
