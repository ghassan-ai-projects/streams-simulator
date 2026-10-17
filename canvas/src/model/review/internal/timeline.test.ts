import { describe, expect, it } from "vitest";
import { report } from "../../../test/runFixture";
import {
  effectorAt,
  eventMarks,
  faultedAt,
  firstEventLine,
  kindsOf,
  nextAlarm,
  nextEvent,
  previousEvent,
  runClock,
} from "./timeline";

const parsed = report;

describe("first event line", () => {
  it("names the first mark with its time and instance, and counts the effector calls", () => {
    const line = firstEventLine(parsed, parsed.startNs);
    expect(line).toMatch(/^First: disturbance · drop on delivery at 00:05:00 · \d+ effector calls?$/);
  });

  it("names a fault as a fault and leaves the count off when the only call is the first event", () => {
    const fault = { ...parsed, disturbances: parsed.disturbances.filter((d) => d.kind === "fault"), effectorCalls: [] };
    expect(firstEventLine(fault, parsed.startNs)).toBe("First: fault · fan failure on CT-1 at 00:10:00");
    const call = { ...parsed, disturbances: [], alarmActivations: [], effectorCalls: parsed.effectorCalls.slice(0, 1) };
    expect(firstEventLine(call, parsed.startNs)).toMatch(/^First: effector call · .+ at \d\d:\d\d:\d\d$/);
  });

  it("says there is nothing when the run has no fault, call or alarm, and nothing before the report", () => {
    const quiet = { ...parsed, disturbances: [], effectorCalls: [], alarmActivations: [] };
    expect(firstEventLine(quiet, parsed.startNs)).toBe("No faults or effector calls in this run");
    expect(firstEventLine(null, 0)).toBeNull();
  });
});

describe("timeline", () => {
  it("shows reported alarm instants as navigation marks without inferring an active interval", () => {
    const alarmActivations = [{ atNs: 650e9, entity: "CH-1", state: "temp_alarm", priority: "high" }];
    const marks = eventMarks({ ...parsed, alarmActivations });
    expect(marks.find((mark) => mark.tone === "alarm")).toEqual({
      atNs: 650e9,
      label: "CH-1 temp alarm activated",
      raw: "alarm activation CH-1 temp_alarm · high",
      kind: "alarm",
      tone: "alarm",
      entity: "CH-1",
      untilNs: null,
    });
    expect(nextEvent(marks, 600e9)).toBe(650e9);
    expect(previousEvent(marks, 900e9)).toBe(650e9);
  });
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
    expect([...faultedAt(parsed, 600e9)]).toEqual([["CT-1", ["fan failure"]]]);
    const fault = parsed.disturbances.find((d) => d.kind === "fault");
    const cleared = {
      ...parsed,
      disturbances: fault === undefined ? [] : [{ ...fault, clearedNs: 700e9, clearedBy: "pulse_end" }],
    };
    expect([...faultedAt(cleared, 699e9)]).toEqual([["CT-1", ["fan failure"]]]);
    expect([...faultedAt(cleared, 700e9)]).toEqual([]);
    const two = {
      ...parsed,
      disturbances: fault === undefined ? [] : [fault, { ...fault, name: "tower__basin_leak" }],
    };
    expect(faultedAt(two, 900e9).get("CT-1")).toEqual(["fan failure", "basin leak"]);
    expect(eventMarks(cleared).find((mark) => mark.atNs === 700e9)?.label).toBe("CT-1 fan failure ended");
    expect([...faultedAt(null, 0)]).toEqual([]);
  });
});

describe("event lane", () => {
  it("reads times from the run's start and names the kinds of event present", () => {
    expect(runClock(5_000e9 + 3_723e9, 5_000e9)).toBe("01:02:03");
    const marks = eventMarks({ ...parsed, alarmActivations: [{ atNs: 1e9, entity: "A", state: "s", priority: "p" }] });
    expect(kindsOf(marks)).toContain("alarm");
    expect(kindsOf([])).toEqual([]);
  });

  it("lists an effector call on its target for 60 s of simulated time after it", () => {
    const calls = [
      { atNs: 100e9, entity: "CP-3", effector: "stop", outcome: "applied · 371 ms" },
      { atNs: 130e9, entity: "CP-3", effector: "start", outcome: "refused: busy" },
    ];
    const run = { ...parsed, effectorCalls: calls };
    expect(effectorAt(run, 99e9).size).toBe(0);
    expect(effectorAt(run, 100e9).get("CP-3")).toBe("stop · applied · 371 ms");
    expect(effectorAt(run, 129e9).get("CP-3")).toBe("stop · applied · 371 ms");
    expect(effectorAt(run, 150e9).get("CP-3")).toBe("start · refused: busy");
    expect(effectorAt(run, 190e9).get("CP-3")).toBe("start · refused: busy");
    expect(effectorAt(run, 191e9).size).toBe(0);
    expect(effectorAt(null, 100e9).size).toBe(0);
  });
});

describe("next alarm", () => {
  const mark = (atNs: number, tone: "alarm" | "fault") =>
    ({ atNs, label: `m${String(atNs)}`, raw: "", kind: "alarm", tone, entity: "CH-1", untilNs: null }) as const;
  const marks = [mark(10, "alarm"), mark(20, "fault"), mark(30, "alarm")];

  it("finds the first alarm after the position, skipping other marks", () => {
    expect(nextAlarm(marks, 10)?.atNs).toBe(30);
    expect(nextAlarm(marks, 5)?.atNs).toBe(10);
  });

  it("wraps to the first alarm after the last, and is none without alarms", () => {
    expect(nextAlarm(marks, 30)?.atNs).toBe(10);
    expect(nextAlarm([mark(20, "fault")], 0)).toBeNull();
  });
});
