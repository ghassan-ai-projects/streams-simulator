import { describe, expect, it } from "vitest";
import { PlaybackIndex } from "./playback";
import { trendFrame, trendOf, trendWindowNs } from "./trend";

const entity = "CT-1";
const channel = "cooling_tower.basin_temp";

function indexOf(samples: readonly (readonly [number, number | string])[]): PlaybackIndex {
  const index = new PlaybackIndex();
  for (const [second, value] of samples) {
    index.add({ atNs: second * 1e9, entity, channel, value, unit: "degC" });
  }
  return index;
}

function every(from: number, to: number, step: number, value: (second: number) => number) {
  const out: (readonly [number, number])[] = [];
  for (let second = from; second <= to; second += step) out.push([second, value(second)]);
  return out;
}

const endNs = 600e9;

describe("the span a trend covers", () => {
  it("starts at the channel's first reading while the run is younger than the window, so the line uses the whole frame", () => {
    const trend = trendOf(indexOf(every(0, 600, 10, (second) => 20 + second / 100)), entity, channel, endNs);
    const xs = trend?.segments.flat().map((point) => point.x) ?? [];
    expect(Math.min(...xs)).toBe(0);
    expect(Math.max(...xs)).toBeCloseTo(trendFrame.width, 6);
    expect(trend?.windowNs).toBe(600e9);
  });

  it("keeps the whole window once the run is older than it", () => {
    const long = 3600e9;
    const trend = trendOf(indexOf(every(0, 3600, 10, () => 20)), entity, channel, long);
    expect(trend?.windowNs).toBe(trendWindowNs);
    expect(Math.min(...(trend?.segments.flat() ?? []).map((point) => point.x))).toBeCloseTo(0, 6);
  });

  it("never narrows below a minute, so a lone fresh reading is a dot at the right edge", () => {
    const trend = trendOf(indexOf([[600, 20]]), entity, channel, endNs);
    expect(trend?.windowNs).toBe(60e9);
    expect(trend?.dots[0]?.x).toBeCloseTo(trendFrame.width, 6);
  });
});

describe("the trend of one channel", () => {
  it("is absent when the channel has no samples in the window or holds words", () => {
    expect(trendOf(indexOf([]), entity, channel, endNs)).toBeNull();
    expect(trendOf(indexOf([[10, 20]]), "nobody", channel, endNs)).toBeNull();
    expect(trendOf(indexOf([[5000, 20]]), entity, channel, endNs)).toBeNull();
    expect(trendOf(indexOf([[10, "off"]]), entity, channel, endNs)).toBeNull();
  });

  it("draws one line over steady readings, scaled to the values it shows and not to zero", () => {
    const trend = trendOf(indexOf(every(0, 600, 10, (second) => 20 + second / 100)), entity, channel, endNs);
    expect(trend?.segments).toHaveLength(1);
    expect(trend?.gaps).toBe(0);
    expect(trend?.min).toBe(20);
    expect(trend?.max).toBe(26);
    const ys = trend?.segments[0]?.map((point) => point.y) ?? [];
    expect(Math.min(...ys)).toBeGreaterThanOrEqual(0);
    expect(Math.max(...ys)).toBeLessThanOrEqual(trendFrame.height);
    expect(ys[0]).toBeGreaterThan(ys.at(-1) ?? 0);
  });

  it("draws a flat run in the middle of the frame", () => {
    const trend = trendOf(indexOf(every(0, 600, 10, () => 21)), entity, channel, endNs);
    expect(new Set(trend?.segments[0]?.map((point) => point.y))).toEqual(new Set([trendFrame.height / 2]));
  });

  it("breaks the line where a gap is longer than three times the median interval and never joins across it", () => {
    const dropout = [...every(0, 200, 10, () => 20), ...every(400, 600, 10, () => 25)];
    const trend = trendOf(indexOf(dropout), entity, channel, endNs);
    expect(trend?.gaps).toBe(1);
    expect(trend?.segments).toHaveLength(2);
    const [before, after] = trend?.segments ?? [];
    const holeStart = (200e9 / endNs) * trendFrame.width;
    const holeEnd = (400e9 / endNs) * trendFrame.width;
    expect(Math.max(...(before ?? []).map((point) => point.x))).toBeLessThanOrEqual(holeStart + 1e-9);
    expect(Math.min(...(after ?? []).map((point) => point.x))).toBeGreaterThanOrEqual(holeEnd - 1e-9);
    expect(trend?.medianNs).toBe(10e9);
  });

  it("does not break the line for a pause shorter than three times the median interval", () => {
    const wobble = [...every(0, 200, 10, () => 20), ...every(225, 600, 10, () => 20)];
    expect(trendOf(indexOf(wobble), entity, channel, endNs)?.gaps).toBe(0);
  });

  it("shows a lone reading as a dot and a pair as a line", () => {
    expect(trendOf(indexOf([[590, 20]]), entity, channel, endNs)).toMatchObject({ segments: [], dots: [{}] });
    expect(
      trendOf(
        indexOf([
          [580, 20],
          [590, 21],
        ]),
        entity,
        channel,
        endNs,
      )?.segments,
    ).toHaveLength(1);
  });

  it("shows an isolated reading after a gap as a dot, not as a line to the earlier ones", () => {
    const trend = trendOf(indexOf([...every(0, 200, 10, () => 20), [590, 22]]), entity, channel, endNs);
    expect(trend?.segments).toHaveLength(1);
    expect(trend?.dots).toHaveLength(1);
    expect(trend?.gaps).toBe(1);
  });

  it("keeps the extremes and the newest value when thousands of readings are thinned", () => {
    const dense = every(0, 600, 0.05, (second) => 20 + Math.sin(second / 7));
    const spiked = [...dense, [300.01, 99] as const, [300.02, -40] as const].sort((a, b) => a[0] - b[0]);
    const trend = trendOf(indexOf(spiked), entity, channel, endNs);
    const points = trend?.segments.flat() ?? [];
    expect(spiked.length).toBeGreaterThan(10_000);
    expect(points.length).toBeLessThan(600);
    expect(trend?.max).toBe(99);
    expect(trend?.min).toBe(-40);
    expect(Math.min(...points.map((point) => point.y))).toBe(0);
    expect(Math.max(...points.map((point) => point.y))).toBe(trendFrame.height);
    expect(points.at(-1)?.x).toBeCloseTo(trendFrame.width, 1);
  });

  it("ends the line at the last reading and leaves the frame empty to its right when the channel went quiet", () => {
    const quiet = trendOf(indexOf(every(0, 300, 10, () => 20)), entity, channel, endNs);
    expect(Math.max(...(quiet?.segments.flat() ?? []).map((point) => point.x))).toBeLessThan(trendFrame.width - 1);
  });
});
