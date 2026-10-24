import { describe, expect, it } from "vitest";
import { PlaybackIndex } from "./playback";
import type { WindowFacts } from "./trendWindow";
import { choiceOf, trendSpanOf, windowWords } from "./trendWindow";

const minute = 60e9;
const channel = { entity: "CT-1", channel: "cooling_tower.basin_temp" };

const facts: WindowFacts = {
  choice: null,
  running: true,
  reachedNs: 30 * minute,
  cursorNs: 12 * minute,
  runStartNs: 0,
  runEndNs: 60 * minute,
};

function indexOf(): PlaybackIndex {
  return new PlaybackIndex();
}

describe("the time window of the trends", () => {
  it("defaults to ten minutes for a running run and to the whole run for a finished one", () => {
    expect(choiceOf(null, { running: true })).toBe("10");
    expect(choiceOf(null, { running: false })).toBe("all");
    expect(choiceOf("30", { running: false })).toBe("30");
  });

  it("ends a running run's window at the reached run clock and not at the cursor", () => {
    const { span } = trendSpanOf(indexOf(), [channel], facts);
    expect(span).toEqual({ startNs: 20 * minute, endNs: 30 * minute });
  });

  it("covers a finished run from its start to its end, whatever the cursor", () => {
    const { span } = trendSpanOf(indexOf(), [channel], { ...facts, running: false });
    expect(span).toEqual({ startNs: 0, endNs: 60 * minute });
  });

  it("ends a finished run's shorter window at the cursor and never starts before the run", () => {
    const shortRun = { ...facts, running: false, choice: "30" as const };
    expect(trendSpanOf(indexOf(), [channel], shortRun).span).toEqual({ startNs: 0, endNs: 12 * minute });
    expect(trendSpanOf(indexOf(), [channel], { ...shortRun, cursorNs: 50 * minute }).span).toEqual({
      startNs: 20 * minute,
      endNs: 50 * minute,
    });
  });

  it("is at least a minute wide so the very start of a run still has an axis", () => {
    const { span } = trendSpanOf(indexOf(), [channel], { ...facts, reachedNs: 10e9 });
    expect(span.endNs - span.startNs).toBe(minute);
  });

  it("starts where a bounded watch stopped keeping readings and says so", () => {
    const index = new PlaybackIndex({ capacity: 4 });
    for (let at = 0; at < 8; at += 1) {
      index.add({
        atNs: (20 + at) * minute,
        entity: channel.entity,
        channel: channel.channel,
        value: at,
        unit: "degC",
      });
    }
    const trimmed = index.trimmedBeforeNs(channel.entity, channel.channel) ?? 0;
    expect(trimmed).toBeGreaterThan(0);
    const limited = trendSpanOf(index, [channel], { ...facts, reachedNs: 30 * minute, choice: "all" });
    expect(limited.limitedFromNs).toBe(trimmed);
    expect(limited.span.startNs).toBe(trimmed);
  });

  it("is not limited when the retention ended before the window starts", () => {
    const index = new PlaybackIndex({ capacity: 4 });
    for (let at = 0; at < 8; at += 1) {
      index.add({ atNs: at * minute, entity: channel.entity, channel: channel.channel, value: at, unit: "degC" });
    }
    const span = trendSpanOf(index, [channel], { ...facts, reachedNs: 40 * minute });
    expect(span.limitedFromNs).toBeNull();
    expect(span.span.startNs).toBe(30 * minute);
  });

  it("describes each window in words for the operator", () => {
    expect(windowWords("10", { running: true })).toBe("10 min of simulated time up to the reached run clock");
    expect(windowWords("30", { running: false })).toBe("30 min of simulated time ending at the cursor");
    expect(windowWords("all", { running: false })).toBe("The whole run");
  });
});
