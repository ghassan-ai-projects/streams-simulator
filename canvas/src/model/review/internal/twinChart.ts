import { runClock } from "./timeline";
import type { TwinDivergence, TwinMetric } from "./twin";
import { numberWords } from "./twinWords";

export const chartFrame = { width: 320, height: 150, left: 44, right: 8, top: 8, bottom: 22 } as const;

interface Mark {
  readonly x: number;
  readonly y: number;
}

interface Shade {
  readonly x: number;
  readonly width: number;
  readonly name: string;
}

interface Tick {
  readonly at: number;
  readonly label: string;
}

export interface TwinChart {
  readonly board: string;
  readonly twin: string;
  readonly boardMarks: readonly Mark[];
  readonly twinMarks: readonly Mark[];
  readonly shades: readonly Shade[];
  readonly xTicks: readonly Tick[];
  readonly yTicks: readonly Tick[];
  readonly cursor: number | null;
}

const markLimit = 40;
const minShade = 3;

const round = (value: number): number => Math.round(value * 10) / 10;

interface Scale {
  readonly x: (ns: number) => number;
  readonly y: (value: number) => number;
  readonly low: number;
  readonly high: number;
}

function scaleOf(points: TwinMetric["series"]["points"], firstNs: number, lastNs: number): Scale {
  const values = points.flatMap((point) => [point.board, point.twin]);
  const low = Math.min(...values);
  const high = Math.max(...values);
  const spread = high === low ? 2 : high - low;
  const bottom = high === low ? low - 1 : low;
  const across = chartFrame.width - chartFrame.left - chartFrame.right;
  const down = chartFrame.height - chartFrame.top - chartFrame.bottom;
  const span = Math.max(lastNs - firstNs, 1);
  return {
    x: (ns) => round(chartFrame.left + (Math.min(Math.max(ns, firstNs), lastNs) - firstNs) * (across / span)),
    y: (value) => round(chartFrame.top + (1 - (value - bottom) / spread) * down),
    low: bottom,
    high: bottom + spread,
  };
}

function line(points: readonly Mark[]): string {
  return points.map((mark) => `${String(mark.x)},${String(mark.y)}`).join(" ");
}

function shadesOf(divergences: readonly TwinDivergence[], scale: Scale): readonly Shade[] {
  return divergences.map((divergence) => {
    const x = scale.x(divergence.firstNs);
    return { x, width: Math.max(round(scale.x(divergence.lastNs) - x), minShade), name: divergence.name };
  });
}

function yTicksOf(scale: Scale): readonly Tick[] {
  const middle = (scale.low + scale.high) / 2;
  return [scale.high, middle, scale.low].map((value) => ({ at: scale.y(value), label: numberWords(value) }));
}

function xTicksOf(scale: Scale, firstNs: number, lastNs: number, startNs: number): readonly Tick[] {
  return [firstNs, lastNs].map((ns) => ({ at: scale.x(ns), label: runClock(ns, startNs) }));
}

const emptyChart: TwinChart = {
  board: "",
  twin: "",
  boardMarks: [],
  twinMarks: [],
  shades: [],
  xTicks: [],
  yTicks: [],
  cursor: null,
};

export function chartOf(
  metric: TwinMetric,
  divergences: readonly TwinDivergence[],
  startNs: number,
  cursorNs: number,
): TwinChart {
  const { points } = metric.series;
  const first = points[0];
  const last = points.at(-1);
  if (first === undefined || last === undefined) return emptyChart;
  const scale = scaleOf(points, first.atNs, last.atNs);
  const board = points.map((point) => ({ x: scale.x(point.atNs), y: scale.y(point.board) }));
  const twin = points.map((point) => ({ x: scale.x(point.atNs), y: scale.y(point.twin) }));
  const marked = points.length <= markLimit;
  const inside = cursorNs >= first.atNs && cursorNs <= last.atNs;
  return {
    board: line(board),
    twin: line(twin),
    boardMarks: marked ? board : [],
    twinMarks: marked ? twin : [],
    shades: shadesOf(divergences, scale),
    xTicks: xTicksOf(scale, first.atNs, last.atNs, startNs),
    yTicks: yTicksOf(scale),
    cursor: inside ? scale.x(cursorNs) : null,
  };
}
