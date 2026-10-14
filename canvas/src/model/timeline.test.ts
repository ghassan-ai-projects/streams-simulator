import { describe, expect, it } from "vitest";
import { report } from "../test/runFixture";
import { parseReport } from "./report";
import { eventMarks, faultedAt, nextEvent, placeChips, previousEvent } from "./timeline";

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
    expect([...faultedAt(parsed, 600e9)]).toEqual([["CT-1", "fan failure"]]);
    const fault = parsed.disturbances.find((d) => d.kind === "fault");
    const cleared = {
      ...parsed,
      disturbances: fault === undefined ? [] : [{ ...fault, clearedNs: 700e9, clearedBy: "pulse_end" }],
    };
    expect([...faultedAt(cleared, 699e9)]).toEqual([["CT-1", "fan failure"]]);
    expect([...faultedAt(cleared, 700e9)]).toEqual([]);
    expect(eventMarks(cleared).find((mark) => mark.atNs === 700e9)?.label).toBe("CT-1 fan failure ended");
    expect([...faultedAt(null, 0)]).toEqual([]);
  });
});

describe("event lane", () => {
  it("puts close events on separate rows and reuses a row once it is free", () => {
    const lane = { width: 100, rows: 2, gap: 4 };
    const rows = (positions: number[]) =>
      placeChips(
        positions,
        positions.map(() => 20),
        lane,
      ).map((chip) => chip.row);
    expect(rows([0, 30, 31])).toEqual([0, 0, 1]);
    expect(rows([0, 1, 2, 3])).toEqual([0, 1, -1, -1]);
    expect(rows([])).toEqual([]);
    expect(placeChips([95], [40], lane)).toEqual([{ left: 60, row: 0 }]);
  });
});
