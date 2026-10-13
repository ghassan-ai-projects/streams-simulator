import { describe, expect, it } from "vitest";
import { report } from "../test/runFixture";
import { parseReport } from "./report";
import { eventMarks, faultedAt, laneRows, nextEvent, previousEvent } from "./timeline";

const parsed = parseReport(report);

describe("timeline", () => {
  it("lists disturbances and effector calls in time order with words for ids", () => {
    const marks = eventMarks(parsed);
    expect(marks.map((mark) => [mark.atNs / 1e9, mark.label, mark.tone])).toEqual([
      [300, "delivery drop", "other"],
      [600, "CT-1 fan failure", "fault"],
      [900, "weather heat wave", "other"],
      [1200, "CH-1 start", "other"],
      [1300, "CH-1 stop", "other"],
      [1400, "CH-1 stop", "other"],
      [1500, "CH-1 stop", "other"],
    ]);
    expect(eventMarks(null)).toEqual([]);
  });

  it("steps to the previous and next event", () => {
    const marks = eventMarks(parsed);
    expect(previousEvent(marks, 650e9)).toBe(600e9);
    expect(nextEvent(marks, 600e9)).toBe(900e9);
    expect(previousEvent(marks, 300e9)).toBeNull();
    expect(nextEvent(marks, 1500e9)).toBeNull();
  });

  it("knows which instances carry an active fault at a time", () => {
    expect([...faultedAt(parsed, 599e9)]).toEqual([]);
    expect([...faultedAt(parsed, 600e9)]).toEqual(["CT-1"]);
    const cleared = {
      ...parsed,
      disturbances: [
        ...parsed.disturbances,
        { kind: "fault", action: "clear", injectedNs: 700e9, onsetNs: 700e9, target: "CT-1", name: "f-0" },
      ],
    } as const;
    expect([...faultedAt(cleared, 800e9)]).toEqual([]);
    expect(eventMarks(cleared).find((mark) => mark.atNs === 700e9)?.label).toBe("CT-1 f-0 cleared");
    expect([...faultedAt(null, 0)]).toEqual([]);
  });
});

describe("event lane", () => {
  it("puts close events on separate rows and reuses a row once it is free", () => {
    expect(laneRows([0, 5, 10, 30, 31], 20)).toEqual([0, 1, 2, 0, 1]);
    expect(laneRows([0, 25, 26], 20, 2)).toEqual([0, 0, 1]);
    expect(laneRows([0, 1, 2, 3], 20, 2)).toEqual([0, 1, -1, -1]);
    expect(laneRows([], 20)).toEqual([]);
  });
});
