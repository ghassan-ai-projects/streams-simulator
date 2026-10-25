import { describe, expect, it } from "vitest";
import { PlaybackIndex } from "./playback";
import type { Quality } from "../../shared/quality";
import { nsAtX, plotFrame, plotOf, xOf } from "./trendPlot";

const basin = { entity: "CT-1", channel: "cooling_tower.basin_temp" };
const power = { entity: "CH-1", channel: "chiller.power" };

interface Sample {
  readonly second: number;
  readonly value: number | string;
  readonly quality?: Quality;
}

function indexOf(samples: readonly Sample[], watched = basin, unit = "degC"): PlaybackIndex {
  const index = new PlaybackIndex();
  for (const sample of samples) {
    index.add({ atNs: sample.second * 1e9, ...watched, value: sample.value, unit, quality: sample.quality ?? "" });
  }
  return index;
}

function every(from: number, to: number, step: number, value: (second: number) => number): Sample[] {
  const out: Sample[] = [];
  for (let second = from; second <= to; second += step) out.push({ second, value: value(second) });
  return out;
}

const span = { startNs: 0, endNs: 600e9 };

describe("one plot of the trends panel", () => {
  it("is empty when the channel has nothing in the window and words when it holds states", () => {
    expect(plotOf(indexOf([]), basin, span).state).toBe("empty");
    expect(plotOf(indexOf([{ second: 5000, value: 20 }]), basin, span).state).toBe("empty");
    expect(plotOf(indexOf([{ second: 5, value: "idle" }]), basin, span).state).toBe("words");
  });

  it("carries its own unit symbol and the code behind it", () => {
    const plot = plotOf(indexOf(every(0, 600, 10, () => 20)), basin, span);
    expect([plot.unit, plot.unitTitle]).toEqual(["°C", "unit code: degC"]);
    const kilowatts = plotOf(
      indexOf(
        every(0, 600, 10, () => 40),
        power,
        "kW",
      ),
      power,
      span,
    );
    expect([kilowatts.unit, kilowatts.unitTitle]).toEqual(["kW", ""]);
  });

  it("puts each extreme's label on the side of the plot where the line is not, so the line's start stays readable", () => {
    const rising = plotOf(indexOf(every(0, 600, 10, (second) => 20 + second / 100)), basin, span);
    expect([rising.top?.side, rising.bottom?.side]).toEqual(["start", "end"]);
    const falling = plotOf(indexOf(every(0, 600, 10, (second) => 26 - second / 100)), basin, span);
    expect([falling.top?.side, falling.bottom?.side]).toEqual(["end", "start"]);
  });

  it("scales y to its own readings, labels the true extremes and keeps them inside the frame", () => {
    const plot = plotOf(indexOf(every(0, 600, 10, (second) => 20 + second / 100)), basin, span);
    expect([plot.top?.text, plot.bottom?.text]).toEqual(["26", "20"]);
    expect(plot.top?.y).toBeGreaterThan(0);
    expect(plot.bottom?.y).toBeLessThan(plotFrame.height);
    expect(plot.top?.y).toBeLessThan(plot.bottom?.y ?? 0);
  });

  it("puts two channels of different units on separate scales that do not influence each other", () => {
    const alone = plotOf(indexOf(every(0, 600, 10, () => 20)), basin, span);
    const withOther = plotOf(indexOf(every(0, 600, 10, () => 20)), basin, span);
    expect(withOther.segments).toEqual(alone.segments);
    const hundreds = plotOf(
      indexOf(
        every(0, 600, 10, (second) => 400 + second),
        power,
        "kW",
      ),
      power,
      span,
    );
    expect(hundreds.top?.text).toBe("1 000");
  });

  it("draws a flat channel in the middle of its frame", () => {
    const plot = plotOf(indexOf(every(0, 600, 10, () => 21)), basin, span);
    expect(new Set(plot.segments[0]?.map((point) => point.y))).toEqual(new Set([plotFrame.height / 2]));
  });

  it("shares one time axis: the same instant has the same x in every plot", () => {
    const first = plotOf(indexOf(every(0, 600, 10, () => 20)), basin, span);
    const second = plotOf(
      indexOf(
        every(0, 600, 10, (s) => s),
        power,
        "kW",
      ),
      power,
      span,
    );
    const at300 = (points: readonly { readonly x: number }[]) => points.find((point) => Math.abs(point.x - 300) < 1e-9);
    expect(at300(first.segments[0] ?? [])).toBeDefined();
    expect(at300(second.segments[0] ?? [])).toBeDefined();
    expect(xOf(300e9, span)).toBe(300);
    expect(nsAtX(300, span)).toBe(300e9);
    expect(nsAtX(-50, span)).toBe(0);
    expect(nsAtX(10_000, span)).toBe(600e9);
  });

  it("leaves a gap longer than three times the median interval blank and never joins across it", () => {
    const dropout = [...every(0, 200, 10, () => 20), ...every(400, 600, 10, () => 25)];
    const plot = plotOf(indexOf(dropout), basin, span);
    expect(plot.gaps).toBe(1);
    expect(plot.segments).toHaveLength(2);
    const [before, after] = plot.segments;
    expect(Math.max(...(before ?? []).map((point) => point.x))).toBeLessThanOrEqual(xOf(200e9, span) + 1e-9);
    expect(Math.min(...(after ?? []).map((point) => point.x))).toBeGreaterThanOrEqual(xOf(400e9, span) - 1e-9);
    expect(plot.medianNs).toBe(10e9);
  });

  it("does not draw a missing reading as zero: the lowest drawn point is the lowest reading", () => {
    const dropout = [...every(0, 100, 10, () => 20), ...every(500, 600, 10, () => 22)];
    const plot = plotOf(indexOf(dropout), basin, span);
    expect(plot.bottom?.text).toBe("20");
    expect(plot.segments.flat().every((point) => point.y >= 0 && point.y <= plotFrame.height)).toBe(true);
  });

  it("thins thousands of readings but keeps the spike and the dip", () => {
    const dense = every(0, 600, 0.05, (second) => 20 + Math.sin(second / 7));
    const spiked = [...dense, { second: 300.01, value: 99 }, { second: 300.02, value: -40 }].sort(
      (a, b) => a.second - b.second,
    );
    const plot = plotOf(indexOf(spiked), basin, span);
    expect(spiked.length).toBeGreaterThan(10_000);
    expect(plot.segments.flat().length).toBeLessThan(1500);
    expect([plot.top?.text, plot.bottom?.text]).toEqual(["99", "-40"]);
  });

  it("marks consecutive suspect readings as one band and keeps stale apart from suspect", () => {
    const marked = every(0, 600, 10, () => 20).map((sample): Sample => {
      if (sample.second >= 100 && sample.second <= 130) return { ...sample, quality: "suspect" };
      return sample.second === 300 ? { ...sample, quality: "stale" } : sample;
    });
    const plot = plotOf(indexOf(marked), basin, span);
    expect(plot.bands.map((band) => [band.quality, band.words])).toEqual([
      ["suspect", "suspect"],
      ["stale", "stale"],
    ]);
    expect(plot.bands[0]?.x0).toBe(xOf(100e9, span));
    expect(plot.bands[0]?.x1).toBe(xOf(130e9, span));
  });

  it("only looks at readings inside the window", () => {
    const plot = plotOf(indexOf(every(0, 600, 10, (second) => second)), basin, { startNs: 300e9, endNs: 600e9 });
    expect([plot.bottom?.text, plot.top?.text]).toEqual(["300", "600"]);
  });
});
