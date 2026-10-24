import type { PlaybackIndex } from "./playback";

export const trendWindowNs = 30 * 60e9;
export const trendFrame = { width: 120, height: 28 } as const;

const gapFactor = 3;
const buckets = 60;

interface TrendPoint {
  readonly x: number;
  readonly y: number;
}

export interface Trend {
  readonly segments: readonly (readonly TrendPoint[])[];
  readonly dots: readonly TrendPoint[];
  readonly min: number;
  readonly max: number;
  readonly gaps: number;
  readonly medianNs: number;
  readonly windowNs: number;
}

interface Run {
  readonly times: readonly number[];
  readonly values: readonly number[];
}

function numeric(values: readonly (number | string | boolean)[]): readonly number[] | null {
  const numbers = values.filter((value): value is number => typeof value === "number" && Number.isFinite(value));
  return numbers.length === values.length ? numbers : null;
}

function medianInterval(times: readonly number[]): number {
  const intervals = times.slice(1).map((time, at) => time - (times[at] ?? time));
  const sorted = intervals.toSorted((a, b) => a - b);
  return sorted[sorted.length >> 1] ?? 0;
}

function split(run: Run, medianNs: number): readonly Run[] {
  const runs: { times: number[]; values: number[] }[] = [{ times: [], values: [] }];
  run.times.forEach((time, at) => {
    const previous = run.times[at - 1];
    if (previous !== undefined && medianNs > 0 && time - previous > gapFactor * medianNs) {
      runs.push({ times: [], values: [] });
    }
    runs.at(-1)?.times.push(time);
    runs.at(-1)?.values.push(run.values[at] ?? 0);
  });
  return runs;
}

function bucketOf(time: number, startNs: number): number {
  return Math.min(buckets - 1, Math.max(0, Math.floor(((time - startNs) / trendWindowNs) * buckets)));
}

function keptIndices(run: Run, startNs: number): readonly number[] {
  if (run.times.length <= buckets * 4) return run.times.map((_, at) => at);
  const kept = new Set<number>([0, run.times.length - 1]);
  const spans = new Map<number, number[]>();
  run.times.forEach((time, at) => {
    const bucket = bucketOf(time, startNs);
    spans.set(bucket, [...(spans.get(bucket) ?? []), at]);
  });
  for (const span of spans.values()) extremes(span, run.values).forEach((at) => kept.add(at));
  return [...kept].sort((a, b) => a - b);
}

function extremes(span: readonly number[], values: readonly number[]): readonly number[] {
  const ranked = span.toSorted((a, b) => (values[a] ?? 0) - (values[b] ?? 0));
  return [span[0] ?? 0, ranked[0] ?? 0, ranked.at(-1) ?? 0, span.at(-1) ?? 0];
}

function plot(run: Run, startNs: number, range: { readonly min: number; readonly max: number }): TrendPoint[] {
  const height = trendFrame.height;
  return keptIndices(run, startNs).map((at) => {
    const time = run.times[at] ?? startNs;
    const value = run.values[at] ?? range.min;
    const level = range.max === range.min ? 0.5 : (value - range.min) / (range.max - range.min);
    return { x: ((time - startNs) / trendWindowNs) * trendFrame.width, y: height - level * height };
  });
}

export function trendOf(index: PlaybackIndex, entity: string, channel: string, endNs: number): Trend | null {
  const startNs = endNs - trendWindowNs;
  const window = index.samples(entity, channel, startNs, endNs);
  const values = window === null ? null : numeric(window.values);
  if (window === null || values === null || values.length === 0) return null;
  const medianNs = medianInterval(window.times);
  const runs = split({ times: window.times, values }, medianNs);
  const range = { min: Math.min(...values), max: Math.max(...values) };
  const plotted = runs.map((run) => plot(run, startNs, range));
  return {
    segments: plotted.filter((points) => points.length > 1),
    dots: plotted.filter((points) => points.length === 1).flat(),
    ...range,
    gaps: runs.length - 1,
    medianNs,
    windowNs: trendWindowNs,
  };
}
