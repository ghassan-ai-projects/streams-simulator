import { describe, expect, it } from "vitest";
import { firstOf, twinModel } from "../../../test/twinReportFixture";
import type { TwinMetric } from "./twin";
import { chartFrame, chartOf } from "./twinChart";

const metric = firstOf(twinModel().metrics);
const divergences = twinModel().divergences;
const withPoints = (points: TwinMetric["series"]["points"]): TwinMetric => ({
  ...metric,
  series: { points, total: points.length, truncated: false },
});

describe("the geometry of the twin chart", () => {
  it("draws the board and the twin over the same scales, inside the plot", () => {
    const chart = chartOf(metric, divergences, 0, 0);
    const xs = chart.board.split(" ").map((pair) => Number(pair.split(",")[0]));
    expect(xs[0]).toBe(chartFrame.left);
    expect(xs.at(-1)).toBe(chartFrame.width - chartFrame.right);
    expect(chart.board.split(" ")).toHaveLength(5);
    expect(chart.twin.split(" ")).toHaveLength(5);
    expect(chart.board.split(" ")[0]).toBe(chart.twin.split(" ")[0]);
    expect(chart.board.split(" ")[4]).not.toBe(chart.twin.split(" ")[4]);
  });

  it("shades each divergence from its first to its last reading and labels the axes with run times and values", () => {
    const chart = chartOf(metric, divergences, 0, 0);
    expect(chart.shades).toHaveLength(1);
    expect(chart.shades[0]?.name).toBe("sensor.temperature#1");
    expect((chart.shades[0]?.x ?? 0) + (chart.shades[0]?.width ?? 0)).toBe(chartFrame.width - chartFrame.right);
    expect(chart.xTicks.map((tick) => tick.label)).toEqual(["00:01:40", "00:08:20"]);
    expect(chart.yTicks.map((tick) => tick.label)).toEqual(["30", "25", "20"]);
  });

  it("gives a stretch of one reading a visible width", () => {
    const one = [{ ...firstOf(divergences), firstNs: 300e9, lastNs: 300e9 }];
    expect(chartOf(metric, one, 0, 0).shades[0]?.width).toBe(3);
  });

  it("marks the readings of a short series, and none of a long one", () => {
    expect(chartOf(metric, [], 0, 0).boardMarks).toHaveLength(5);
    expect(chartOf(metric, [], 0, 0).twinMarks).toHaveLength(5);
    const long = Array.from({ length: 41 }, (_, index) => ({ atNs: index * 1e9, board: index, twin: index }));
    const chart = chartOf(withPoints(long), [], 0, 0);
    expect(chart.boardMarks).toEqual([]);
    expect(chart.twinMarks).toEqual([]);
  });

  it("puts the cursor on the chart only when the playback time lies inside the series", () => {
    expect(chartOf(metric, [], 0, 300e9).cursor).toBe(
      chartFrame.left + (chartFrame.width - chartFrame.left - chartFrame.right) / 2,
    );
    expect(chartOf(metric, [], 0, 50e9).cursor).toBeNull();
    expect(chartOf(metric, [], 0, 900e9).cursor).toBeNull();
  });

  it("draws a flat series in the middle of its scale and a single reading at the left edge", () => {
    const flat = chartOf(withPoints([{ atNs: 1e9, board: 20, twin: 20 }]), [], 0, 0);
    expect(flat.board).toBe(`${String(chartFrame.left)},${String(flat.board.split(",")[1])}`);
    expect(flat.yTicks.map((tick) => tick.label)).toEqual(["21", "20", "19"]);
    expect(flat.xTicks).toHaveLength(2);
  });

  it("draws nothing for a series with no point", () => {
    expect(chartOf(withPoints([]), divergences, 0, 0)).toMatchObject({ board: "", twin: "", shades: [], cursor: null });
  });
});
